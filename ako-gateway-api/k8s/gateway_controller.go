/*
 * Copyright 2023-2024 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package k8s

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	gatewayexternalversions "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var controllerInstance *GatewayController
var ctrlonce sync.Once

type GatewayController struct {
	worker_id   uint32
	informers   *utils.Informers
	workqueue   []workqueue.RateLimitingInterface
	DisableSync bool
}

func SharedGatewayController() *GatewayController {
	ctrlonce.Do(func() {
		controllerInstance = &GatewayController{
			worker_id:   (uint32(1) << utils.NumWorkersIngestion) - 1,
			informers:   utils.GetInformers(),
			DisableSync: true,
		}
	})
	return controllerInstance
}

func (c *GatewayController) InitGatewayAPIInformers(cs gatewayclientset.Interface) {
	gatewayFactory := gatewayexternalversions.NewSharedInformerFactory(cs, time.Second*30)
	akogatewayapilib.AKOControlConfig().SetGatewayApiInformers(&akogatewayapilib.GatewayAPIInformers{
		GatewayInformer:      gatewayFactory.Gateway().V1beta1().Gateways(),
		GatewayClassInformer: gatewayFactory.Gateway().V1beta1().GatewayClasses(),
		HTTPRouteInformer:    gatewayFactory.Gateway().V1beta1().HTTPRoutes(),
	})
}

func (c *GatewayController) Start(stopCh <-chan struct{}) {
	go c.informers.ServiceInformer.Informer().Run(stopCh)
	go c.informers.EpInformer.Informer().Run(stopCh)
	go c.informers.NSInformer.Informer().Run(stopCh)

	informersList := []cache.InformerSynced{
		c.informers.EpInformer.Informer().HasSynced,
		c.informers.ServiceInformer.Informer().HasSynced,
		c.informers.NSInformer.Informer().HasSynced,
	}

	if !lib.AviSecretInitialized {
		go c.informers.SecretInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.SecretInformer.Informer().HasSynced)
	}

	if lib.GetServiceType() == lib.NodePortLocal {
		go c.informers.PodInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.PodInformer.Informer().HasSynced)
	}

	go akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Informer().Run(stopCh)
	informersList = append(informersList, akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Informer().HasSynced)
	go akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Informer().Run(stopCh)
	informersList = append(informersList, akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Informer().HasSynced)
	go akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Informer().Run(stopCh)
	informersList = append(informersList, akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Informer().HasSynced)

	if !cache.WaitForCacheSync(stopCh, informersList...) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced")
	}
}

func (c *GatewayController) SetupEventHandlers(k8sinfo k8s.K8sinformers) {
	mcpQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	c.workqueue = mcpQueue.Workqueue
	numWorkers := mcpQueue.NumWorkers

	epEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ep := obj.(*corev1.Endpoints)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ep))
			key := utils.Endpoints + "/" + utils.ObjKey(ep)
			if lib.IsNamespaceBlocked(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Endpoint Add event: Namespace: %s didn't qualify filter", key, namespace)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ep, ok := obj.(*corev1.Endpoints)
			if !ok {
				// endpoints were deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				ep, ok = tombstone.Obj.(*corev1.Endpoints)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Endpoints: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ep))
			key := utils.Endpoints + "/" + utils.ObjKey(ep)
			if lib.IsNamespaceBlocked(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Endpoint Update event: Namespace: %s didn't qualify filter", key, namespace)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oep := old.(*corev1.Endpoints)
			cep := cur.(*corev1.Endpoints)
			if !reflect.DeepEqual(cep.Subsets, oep.Subsets) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(cep))
				key := utils.Endpoints + "/" + utils.ObjKey(cep)
				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Endpoint Update event: Namespace: %s didn't qualify filter", key, namespace)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}
	c.informers.EpInformer.Informer().AddEventHandler(epEventHandler)

	svcEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			svc := obj.(*corev1.Service)
			key := utils.Service + "/" + utils.ObjKey(svc)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == svc.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: same resource version returning", key)
				return
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			svc, ok := obj.(*corev1.Service)
			if !ok {
				// service was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				svc, ok = tombstone.Obj.(*corev1.Service)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Service: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			key := utils.Service + "/" + utils.ObjKey(svc)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			objects.SharedResourceVerInstanceLister().Delete(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldobj := old.(*corev1.Service)
			svc := cur.(*corev1.Service)
			if oldobj.ResourceVersion != svc.ResourceVersion || !reflect.DeepEqual(svc.Annotations, oldobj.Annotations) {
				// Only add the key if the resource versions have changed.
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
				key := utils.Service + "/" + utils.ObjKey(svc)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}
	c.informers.ServiceInformer.Informer().AddEventHandler(svcEventHandler)

	secretEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			secret := obj.(*corev1.Secret)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
			key := utils.Secret + "/" + utils.ObjKey(secret)
			if lib.IsNamespaceBlocked(namespace) {
				utils.AviLog.Debugf("key: %s, msg: secret add event. namespace: %s didn't qualify filter", key, namespace)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				secret, ok = tombstone.Obj.(*corev1.Secret)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a Secret: %#v", obj)
					return
				}
			}
			if checkAviSecretUpdateAndShutdown(secret) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
				key := utils.Secret + "/" + utils.ObjKey(secret)
				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key: %s, msg: secret delete event. namespace: %s didn't qualify filter", key, namespace)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldobj := old.(*corev1.Secret)
			secret := cur.(*corev1.Secret)
			if oldobj.ResourceVersion != secret.ResourceVersion && !reflect.DeepEqual(secret.Data, oldobj.Data) {
				if checkAviSecretUpdateAndShutdown(secret) {
					// Only add the key if the resource versions have changed.
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
					key := utils.Secret + "/" + utils.ObjKey(secret)
					if lib.IsNamespaceBlocked(namespace) {
						utils.AviLog.Debugf("key: %s, msg: secret update event. namespace: %s didn't qualify filter", key, namespace)
						return
					}
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				}
			}
		},
	}
	if c.informers.SecretInformer != nil {
		c.informers.SecretInformer.Informer().AddEventHandler(secretEventHandler)
	}

	//Add namespace event handler if migration is enabled and informer not nil
	nsFilterObj := utils.GetGlobalNSFilter()
	if nsFilterObj.EnableMigration && c.informers.NSInformer != nil {
		utils.AviLog.Debug("Adding namespace event handler")
		namespaceEventHandler := AddNamespaceEventHandler(numWorkers, c)
		c.informers.NSInformer.Informer().AddEventHandler(namespaceEventHandler)
	}

}

func AddNamespaceEventHandler(numWorkers uint32, c *GatewayController) cache.ResourceEventHandler {
	namespaceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ns := obj.(*corev1.Namespace)
			nsLabels := ns.GetLabels()
			namespace := ns.GetName()
			if utils.CheckIfNamespaceAccepted(namespace, nsLabels, false) {
				utils.AddNamespaceToFilter(ns.GetName())
				utils.AviLog.Debugf("NS Add event: Namespace passed filter: %s", ns.GetName())
			} else {
				//Case: previously deleted valid NS, added back with no labels or invalid labels but nsList contain that ns
				utils.AviLog.Debugf("NS Add event: Namespace did not pass filter: %s", ns.GetName())
				if utils.CheckIfNamespaceAccepted(namespace) {
					utils.AviLog.Debugf("Ns Add event: Deleting previous valid namespace: %s from valid NS List", ns.GetName())
					utils.DeleteNamespaceFromFilter(ns.GetName())
				}
			}

		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			nsOld := old.(*corev1.Namespace)
			nsCur := cur.(*corev1.Namespace)
			if isNamespaceUpdated(nsOld, nsCur) {
				oldNSAccepted := utils.CheckIfNamespaceAccepted(nsOld.GetName(), nsOld.Labels, false)
				newNSAccepted := utils.CheckIfNamespaceAccepted(nsCur.GetName(), nsCur.Labels, false)

				if !oldNSAccepted && newNSAccepted {
					//Case 1: Namespace updated with valid labels
					//Call ingress/route and service add
					utils.AddNamespaceToFilter(nsCur.GetName())
					if utils.GetInformers().ServiceInformer != nil {
						utils.AviLog.Debugf("Adding L4 services for namespaces: %s", nsCur.GetName())
						AddServicesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
					if lib.UseServicesAPI() {
						utils.AviLog.Debugf("Adding Gatways for namespaces: %s", nsCur.GetName())
						AddGatewaysFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
				} else if oldNSAccepted && !newNSAccepted {
					//Case 2: Old valid namespace updated with invalid labels
					//Call ingress/route and service delete
					utils.DeleteNamespaceFromFilter(nsCur.GetName())
					if utils.GetInformers().ServiceInformer != nil {
						utils.AviLog.Debugf("Deleting L4 services for namespaces: %s", nsCur.GetName())
						AddServicesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					}
					if lib.UseServicesAPI() {
						utils.AviLog.Debugf("Deleting Gatways for namespaces: %s", nsCur.GetName())
						AddGatewaysFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					}
				}
			}
		},
	}
	return namespaceEventHandler
}
func isNamespaceUpdated(oldNS, newNS *corev1.Namespace) bool {
	if oldNS.ResourceVersion == newNS.ResourceVersion {
		return false
	}
	oldLabelHash := utils.Hash(utils.Stringify(oldNS.Labels))
	newLabelHash := utils.Hash(utils.Stringify(newNS.Labels))
	return oldLabelHash != newLabelHash
}

func AddServicesFromNSToIngestionQueue(numWorkers uint32, c *GatewayController, namespace string, msg string) {
	var key string
	svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the services during namespace sync: %s", err)
		return
	}
	for _, svcObj := range svcObjs {

		key = utils.Service + "/" + utils.ObjKey(svcObj)

		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}
}

func AddGatewaysFromNSToIngestionQueue(numWorkers uint32, c *GatewayController, namespace string, msg string) {
	gwObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("NS to gateways queue add: Error occurred while retrieving gateways for namespace: %s", namespace)
		return
	}
	for _, gwObj := range gwObjs {
		key := lib.Gateway + "/" + utils.ObjKey(gwObj)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}

	gwClassObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("NS to gateways queue add: Error occurred while retrieving gateway class for namespace: %s", namespace)
		return
	}
	for _, gwClassObj := range gwClassObjs {
		key := lib.GatewayClass + "/" + utils.ObjKey(gwClassObj)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}
}

func checkAviSecretUpdateAndShutdown(secret *corev1.Secret) bool {
	if secret.Namespace == utils.GetAKONamespace() && secret.Name == lib.AviSecret {
		// if the secret is updated or deleted we shutdown API server
		utils.AviLog.Warnf("Avi Secret object %s/%s updated/deleted, shutting down AKO", secret.Namespace, secret.Name)
		lib.ShutdownApi()
		return false
	}
	return true
}

func (c *GatewayController) SetupGatewayApiEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up Gateway API Event handlers")
	informer := akogatewayapilib.AKOControlConfig().GatewayApiInformers()

	gatewayEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gw := obj.(*gatewayv1beta1.Gateway)
			key := lib.Gateway + "/" + utils.ObjKey(gw)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == gw.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: same resource version returning", key)
				return
			}
			if !IsValidGateway(key, gw) {
				return
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gw, ok := obj.(*gatewayv1beta1.Gateway)
			if !ok {
				// gateway was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				gw, ok = tombstone.Obj.(*gatewayv1beta1.Gateway)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a Gateway: %#v", obj)
					return
				}
			}
			key := lib.Gateway + "/" + utils.ObjKey(gw)
			objects.SharedResourceVerInstanceLister().Delete(key)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, obj interface{}) {
			if c.DisableSync {
				return
			}
			oldGw := old.(*gatewayv1beta1.Gateway)
			gw := obj.(*gatewayv1beta1.Gateway)
			if IsGatewayUpdated(oldGw, gw) {
				key := lib.Gateway + "/" + utils.ObjKey(gw)
				if !IsValidGateway(key, gw) {
					return
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}
	informer.GatewayInformer.Informer().AddEventHandler(gatewayEventHandler)

	gatewayClassEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gwClass := obj.(*gatewayv1beta1.GatewayClass)
			key := lib.GatewayClass + "/" + utils.ObjKey(gwClass)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == gwClass.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: same resource version returning", key)
				return
			}
			if !IsGatewayClassValid(key, gwClass) {
				return
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwClass))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gwClass, ok := obj.(*gatewayv1beta1.GatewayClass)
			if !ok {
				// gateway class was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				gwClass, ok = tombstone.Obj.(*gatewayv1beta1.GatewayClass)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a GatewayClass: %#v", obj)
					return
				}
			}
			controllerName := string(gwClass.Spec.ControllerName)
			if !akogatewayapilib.CheckGatewayClassController(controllerName) {
				return
			}
			key := lib.GatewayClass + "/" + utils.ObjKey(gwClass)
			objects.SharedResourceVerInstanceLister().Delete(key)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwClass))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, obj interface{}) {
			if c.DisableSync {
				return
			}
			oldGwClass := old.(*gatewayv1beta1.GatewayClass)
			gwClass := obj.(*gatewayv1beta1.GatewayClass)
			if !reflect.DeepEqual(oldGwClass.Spec, gwClass.Spec) || gwClass.GetDeletionTimestamp() != nil {
				key := lib.GatewayClass + "/" + utils.ObjKey(gwClass)
				if !IsGatewayClassValid(key, gwClass) {
					return
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwClass))
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}
	informer.GatewayClassInformer.Informer().AddEventHandler(gatewayClassEventHandler)

	httpRouteEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			httpRoute := obj.(*gatewayv1beta1.HTTPRoute)
			key := lib.HTTPRoute + "/" + utils.ObjKey(httpRoute)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == httpRoute.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: same resource version returning", key)
				return
			}
			if !IsHTTPRouteValid(key, httpRoute) {
				return
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httpRoute))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			httpRoute, ok := obj.(*gatewayv1beta1.HTTPRoute)
			if !ok {
				// httpRoute was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				httpRoute, ok = tombstone.Obj.(*gatewayv1beta1.HTTPRoute)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an HTTPRoute: %#v", obj)
					return
				}
			}
			key := lib.HTTPRoute + "/" + utils.ObjKey(httpRoute)
			objects.SharedResourceVerInstanceLister().Delete(key)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httpRoute))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, obj interface{}) {
			if c.DisableSync {
				return
			}
			oldHTTPRoute := old.(*gatewayv1beta1.HTTPRoute)
			newHTTPRoute := obj.(*gatewayv1beta1.HTTPRoute)
			if IsHTTPRouteUpdated(oldHTTPRoute, newHTTPRoute) {
				key := lib.HTTPRoute + "/" + utils.ObjKey(newHTTPRoute)
				if !IsHTTPRouteValid(key, newHTTPRoute) {
					return
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(newHTTPRoute))
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}
	informer.HTTPRouteInformer.Informer().AddEventHandler(httpRouteEventHandler)
}

func IsGatewayUpdated(oldGateway, newGateway *gatewayv1beta1.Gateway) bool {
	if newGateway.GetDeletionTimestamp() != nil {
		return true
	}
	oldHash := utils.Hash(utils.Stringify(oldGateway.Spec))
	newHash := utils.Hash(utils.Stringify(newGateway.Spec))
	return oldHash != newHash
}

func IsHTTPRouteUpdated(oldHTTPRoute, newHTTPRoute *gatewayv1beta1.HTTPRoute) bool {
	if newHTTPRoute.GetDeletionTimestamp() != nil {
		return true
	}
	oldHash := utils.Hash(utils.Stringify(oldHTTPRoute.Spec))
	newHash := utils.Hash(utils.Stringify(newHTTPRoute.Spec))
	return oldHash != newHash
}

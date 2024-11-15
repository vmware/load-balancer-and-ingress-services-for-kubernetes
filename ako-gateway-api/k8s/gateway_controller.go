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
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	gatewayexternalversions "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
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

type keyData struct {
	key string
	ctx context.Context
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
		GatewayInformer:      gatewayFactory.Gateway().V1().Gateways(),
		GatewayClassInformer: gatewayFactory.Gateway().V1().GatewayClasses(),
		HTTPRouteInformer:    gatewayFactory.Gateway().V1().HTTPRoutes(),
	})
}

func (c *GatewayController) Start(stopCh <-chan struct{}) {
	go c.informers.ServiceInformer.Informer().Run(stopCh)

	informersList := []cache.InformerSynced{
		c.informers.ServiceInformer.Informer().HasSynced,
	}

	if lib.AKOControlConfig().GetEndpointSlicesEnabled() {
		go c.informers.EpSlicesInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.EpSlicesInformer.Informer().HasSynced)
	} else if lib.GetServiceType() == lib.NodePortLocal {
		go c.informers.PodInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.PodInformer.Informer().HasSynced)
	} else {
		go c.informers.EpInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.EpInformer.Informer().HasSynced)
	}

	if !lib.AviSecretInitialized {
		go c.informers.SecretInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.SecretInformer.Informer().HasSynced)
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
		utils.AviLog.Infof("Caches synced")
	}
}

func (c *GatewayController) SetupEventHandlers(k8sinfo k8s.K8sinformers) {
	mcpQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	c.workqueue = mcpQueue.Workqueue
	numWorkers := mcpQueue.NumWorkers

	// Add EPSInformer
	if lib.AKOControlConfig().GetEndpointSlicesEnabled() {
		epsEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				eps := obj.(*discovery.EndpointSlice)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(eps))
				svcName, ok := eps.Labels[discovery.LabelServiceName]
				if !ok || svcName == "" {
					utils.AviLog.Debugf("Endpointslice Add event: Endpointslice does not have backing svc")
					return
				}

				key := utils.Endpointslices + "/" + namespace + "/" + svcName

				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				eps, ok := obj.(*discovery.EndpointSlice)
				if !ok {
					// endpoints were deleted but its final state is unrecorded.
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					eps, ok = tombstone.Obj.(*discovery.EndpointSlice)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an Endpointslice: %#v", obj)
						return
					}
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(eps))
				svcName, ok := eps.Labels[discovery.LabelServiceName]
				if !ok || svcName == "" {
					utils.AviLog.Debugf("Endpointslice Delete event: Endpointslice does not have backing svc")
					return
				}

				key := utils.Endpointslices + "/" + namespace + "/" + svcName

				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			},
			UpdateFunc: func(old, cur interface{}) {
				if c.DisableSync {
					return
				}
				oldEndpointSlice := old.(*discovery.EndpointSlice)
				currentEndpointSlice := cur.(*discovery.EndpointSlice)
				if oldEndpointSlice.ResourceVersion != currentEndpointSlice.ResourceVersion {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(currentEndpointSlice))
					svcName, ok := currentEndpointSlice.Labels[discovery.LabelServiceName]
					if !ok || svcName == "" {
						svcNameOld, ok := oldEndpointSlice.Labels[discovery.LabelServiceName]
						if !ok || svcNameOld == "" {
							utils.AviLog.Debugf("Endpointslice Update event: Endpointslice does not have backing svc")
							return
						}
						svcName = svcNameOld
					}

					key := utils.Endpointslices + "/" + namespace + "/" + svcName

					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				}
			},
		}
		c.informers.EpSlicesInformer.Informer().AddEventHandler(epsEventHandler)
	} else if lib.GetServiceType() == lib.NodePortLocal {
		podEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				pod := obj.(*corev1.Pod)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(pod))
				key := utils.Pod + "/" + utils.ObjKey(pod)
				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key : %s, msg: Pod Add event: Namespace: %s didn't qualify filter", key, namespace)
					return
				}
				ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
				if ok && resVer.(string) == pod.ResourceVersion {
					utils.AviLog.Debugf("key : %s, msg: same resource version returning", key)
					return
				}
				if _, ok := pod.GetAnnotations()[lib.NPLPodAnnotation]; !ok {
					utils.AviLog.Warnf("key : %s, msg: 'nodeportlocal.antrea.io' annotation not found, ignoring the pod", key)
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
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					pod, ok = tombstone.Obj.(*corev1.Pod)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not a Pod: %#v", obj)
						return
					}
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(pod))
				key := utils.Pod + "/" + utils.ObjKey(pod)

				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Pod Delete event: Namespace: %s didn't qualify filter", key, namespace)
					return
				}
				if _, ok := pod.GetAnnotations()[lib.NPLPodAnnotation]; !ok {
					utils.AviLog.Warnf("key : %s, msg: 'nodeportlocal.antrea.io' annotation not found, ignoring the pod", key)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				objects.SharedResourceVerInstanceLister().Delete(key)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			},
			UpdateFunc: func(old, cur interface{}) {
				if c.DisableSync {
					return
				}
				oldPod := old.(*corev1.Pod)
				newPod := cur.(*corev1.Pod)
				key := utils.Pod + "/" + utils.ObjKey(oldPod)

				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(newPod))
				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Pod Update event: Namespace: %s didn't qualify filter", key, namespace)
					return
				}
				if _, ok := newPod.GetAnnotations()[lib.NPLPodAnnotation]; !ok {
					utils.AviLog.Warnf("key : %s, msg: 'nodeportlocal.antrea.io' annotation not found, ignoring the pod", key)
					return
				}
				for _, container := range newPod.Status.ContainerStatuses {
					if !container.Ready {
						if container.State.Terminated != nil {
							utils.AviLog.Warnf("key : %s, msg: Container %s is in terminated state, ignoring pod update", key, container.Name)
							return
						}
						if container.State.Waiting != nil && container.State.Waiting.Reason == "CrashLoopBackOff" {
							utils.AviLog.Warnf("key : %s, msg: Container %s is in CrashLoopBackOff state, ignoring pod update", key, container.Name)
							return
						}
					}
				}
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			},
		}
		c.informers.PodInformer.Informer().AddEventHandler(podEventHandler)
	} else {
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
	}

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
			if oldobj.ResourceVersion != svc.ResourceVersion {
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
			gw := obj.(*gatewayv1.Gateway)

			key := lib.Gateway + "/" + utils.ObjKey(gw)

			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == gw.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: same resource version returning", key)
				return
			}
			valid, allowedRoutesAll := IsValidGateway(key, gw)
			if !valid {
				return
			}
			listRoutes, err := validateReferredHTTPRoute(key, allowedRoutesAll, gw)
			if err != nil {
				utils.AviLog.Errorf("Validation of Referred HTTPRoutes failed due to error : %s", err.Error())
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			for _, route := range listRoutes {
				key := lib.HTTPRoute + "/" + utils.ObjKey(route)

				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gw, ok := obj.(*gatewayv1.Gateway)
			if !ok {
				// gateway was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				gw, ok = tombstone.Obj.(*gatewayv1.Gateway)
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
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayToGatewayStatusMapping(utils.ObjKey(gw))
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, obj interface{}) {
			if c.DisableSync {
				return
			}
			oldGw := old.(*gatewayv1.Gateway)
			gw := obj.(*gatewayv1.Gateway)
			if IsGatewayUpdated(oldGw, gw) {

				key := lib.Gateway + "/" + utils.ObjKey(gw)

				valid, allowedRoutesAll := IsValidGateway(key, gw)
				if !valid {
					return
				}
				listRoutes, err := validateReferredHTTPRoute(key, allowedRoutesAll, gw)
				if err != nil {
					utils.AviLog.Errorf("Validation of Referred HTTPRoutes Failed due to error : %s", err.Error())
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				for _, route := range listRoutes {
					key := lib.HTTPRoute + "/" + utils.ObjKey(route)

					c.workqueue[bkt].AddRateLimited(key)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				}
			}
		},
	}
	informer.GatewayInformer.Informer().AddEventHandler(gatewayEventHandler)

	gatewayClassEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gwClass := obj.(*gatewayv1.GatewayClass)

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
			gwClass, ok := obj.(*gatewayv1.GatewayClass)
			if !ok {
				// gateway class was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				gwClass, ok = tombstone.Obj.(*gatewayv1.GatewayClass)
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
			oldGwClass := old.(*gatewayv1.GatewayClass)
			gwClass := obj.(*gatewayv1.GatewayClass)
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
			httpRoute := obj.(*gatewayv1.HTTPRoute)

			key := lib.HTTPRoute + "/" + utils.ObjKey(httpRoute)

			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == httpRoute.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: same resource version returning", key)
				return
			}
			if !IsHTTPRouteConfigValid(key, httpRoute) {
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
			httpRoute, ok := obj.(*gatewayv1.HTTPRoute)
			if !ok {
				// httpRoute was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				httpRoute, ok = tombstone.Obj.(*gatewayv1.HTTPRoute)
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
			akogatewayapiobjects.GatewayApiLister().DeleteRouteToRouteStatusMapping(utils.ObjKey(httpRoute))
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, obj interface{}) {
			if c.DisableSync {
				return
			}
			oldHTTPRoute := old.(*gatewayv1.HTTPRoute)
			newHTTPRoute := obj.(*gatewayv1.HTTPRoute)
			if IsHTTPRouteUpdated(oldHTTPRoute, newHTTPRoute) {

				key := lib.HTTPRoute + "/" + utils.ObjKey(newHTTPRoute)

				if !IsHTTPRouteConfigValid(key, newHTTPRoute) {
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

func IsGatewayUpdated(oldGateway, newGateway *gatewayv1.Gateway) bool {
	if newGateway.GetDeletionTimestamp() != nil {
		return true
	}
	oldHash := utils.Hash(utils.Stringify(oldGateway.Spec))
	newHash := utils.Hash(utils.Stringify(newGateway.Spec))
	return oldHash != newHash
}

func IsHTTPRouteUpdated(oldHTTPRoute, newHTTPRoute *gatewayv1.HTTPRoute) bool {
	if newHTTPRoute.GetDeletionTimestamp() != nil {
		return true
	}
	oldHash := utils.Hash(utils.Stringify(oldHTTPRoute.Spec))
	newHash := utils.Hash(utils.Stringify(newHTTPRoute.Spec))
	return oldHash != newHash
}

func validateAviConfigMap(obj interface{}) (*corev1.ConfigMap, bool) {
	configMap, ok := obj.(*corev1.ConfigMap)
	if ok && configMap.Namespace == utils.GetAKONamespace() && configMap.Name == lib.AviConfigMap {
		return configMap, true
	}
	return nil, false
}
func validateReferredHTTPRoute(key string, allowedRoutesAll bool, gateway *gatewayv1.Gateway) ([]*gatewayv1.HTTPRoute, error) {
	namespace := gateway.Namespace
	if allowedRoutesAll {
		namespace = metav1.NamespaceAll
	}
	hrObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).List(labels.Set(nil).AsSelector())
	httpRoutes := make([]*gatewayv1.HTTPRoute, 0)
	if err != nil {
		return nil, err
	}
	for _, httpRoute := range hrObjs {
		for _, parentRef := range httpRoute.Spec.ParentRefs {
			if parentRef.Name == gatewayv1.ObjectName(gateway.Name) {
				if IsHTTPRouteConfigValid(key, httpRoute) {
					httpRoutes = append(httpRoutes, httpRoute)
				}
				break
			}
		}
	}
	sort.Slice(httpRoutes, func(i, j int) bool {
		if httpRoutes[i].GetCreationTimestamp().Unix() == httpRoutes[j].GetCreationTimestamp().Unix() {
			return httpRoutes[i].Namespace+"/"+httpRoutes[i].Name < httpRoutes[j].Namespace+"/"+httpRoutes[j].Name
		}
		return httpRoutes[i].GetCreationTimestamp().Unix() < httpRoutes[j].GetCreationTimestamp().Unix()
	})
	return httpRoutes, nil
}

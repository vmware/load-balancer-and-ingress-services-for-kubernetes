/*
 * Copyright 2019-2020 VMware, Inc.
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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	oshiftclient "github.com/openshift/client-go/route/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

var controllerInstance *AviController
var ctrlonce sync.Once

// These tags below are only applicable in case of advanced L4 features at the moment.

// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gateways;gateways/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gatewayclasses;gatewayclasses/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=services;services/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch

type AviController struct {
	worker_id uint32
	//recorder        record.EventRecorder
	informers        *utils.Informers
	dynamicInformers *lib.DynamicInformers
	workqueue        []workqueue.RateLimitingInterface
	DisableSync      bool
}

type K8sinformers struct {
	Cs            kubernetes.Interface
	DynamicClient dynamic.Interface
	OshiftClient  oshiftclient.Interface
}

func SharedAviController() *AviController {
	ctrlonce.Do(func() {
		controllerInstance = &AviController{
			worker_id: (uint32(1) << utils.NumWorkersIngestion) - 1,
			//recorder:  recorder,
			informers:        utils.GetInformers(),
			dynamicInformers: lib.GetDynamicInformers(),
			DisableSync:      true,
		}
	})
	return controllerInstance
}

func isNodeUpdated(oldNode, newNode *corev1.Node) bool {
	if oldNode.ResourceVersion == newNode.ResourceVersion {
		return false
	}
	var oldaddr, newaddr string

	oldAddrs := oldNode.Status.Addresses
	newAddrs := newNode.Status.Addresses
	if len(oldAddrs) != len(newAddrs) {
		return true
	}

	for _, addr := range oldAddrs {
		if addr.Type == "InternalIP" {
			oldaddr = addr.Address
			break
		}
	}
	for _, addr := range newAddrs {
		if addr.Type == "InternalIP" {
			newaddr = addr.Address
			break
		}
	}
	if oldaddr != newaddr {
		return true
	}
	if oldNode.Spec.PodCIDR != newNode.Spec.PodCIDR {
		return true
	}

	nodeLabelEq := reflect.DeepEqual(oldNode.ObjectMeta.Labels, newNode.ObjectMeta.Labels)
	if !nodeLabelEq {
		return true
	}

	return false
}

// Consider an ingress has been updated only if spec/annotation is updated
func isIngressUpdated(oldIngress, newIngress *networkingv1beta1.Ingress) bool {
	if oldIngress.ResourceVersion == newIngress.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldIngress.Spec))
	oldAnnotationHash := utils.Hash(utils.Stringify(oldIngress.Annotations))
	newSpecHash := utils.Hash(utils.Stringify(newIngress.Spec))
	newAnnotationHash := utils.Hash(utils.Stringify(newIngress.Annotations))

	if oldSpecHash != newSpecHash || oldAnnotationHash != newAnnotationHash {
		return true
	}

	return false
}

// Consider a route has been updated only if spec/annotation is updated
func isRouteUpdated(oldRoute, newRoute *routev1.Route) bool {
	if oldRoute.ResourceVersion == newRoute.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldRoute.Spec))
	newSpecHash := utils.Hash(utils.Stringify(newRoute.Spec))

	if oldSpecHash != newSpecHash {
		return true
	}

	return false
}

func isNamespaceUpdated(oldNS, newNS *corev1.Namespace) bool {
	if oldNS.ResourceVersion == newNS.ResourceVersion {
		return false
	}
	oldLabelHash := utils.Hash(utils.Stringify(oldNS.Labels))
	newLabelHash := utils.Hash(utils.Stringify(newNS.Labels))
	if oldLabelHash != newLabelHash {
		return true
	}
	return false
}
func AddIngressFromNSToIngestionQueue(numWorkers uint32, c *AviController, namespace string, msg string) {
	ingObjs, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("NS to ingress queue add: Error occurred while retrieving ingresss for namespace: %s", namespace)
		return
	}
	for _, ingObj := range ingObjs {
		key := utils.Ingress + "/" + utils.ObjKey(ingObj)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}

}
func AddRoutesFromNSToIngestionQueue(numWorkers uint32, c *AviController, namespace string, msg string) {
	routeObjs, err := utils.GetInformers().RouteInformer.Lister().Routes(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("NS to route queue add: Error occurred while retrieving routes for namespace: %s", namespace)
		return
	}
	for _, routeObj := range routeObjs {
		key := utils.OshiftRoute + "/" + utils.ObjKey(routeObj)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}

}

/*
 * Namespace Add event: will be called during each boot or newNS added. In add event
 * handler, just add valid namespaces as Ingress handling, present in namespace, will be done
 * during fullk8sync for reboot case or individual ingress handler called for ingress add event
 * (For second case, user will add namespace first and then ingress, so just validating namespace
 * should be enough)
 */
/* Namespace Delete event: no op. Let individual event handler take care
 */
/* Namespace update event:  2 cases to handle : NS label changed from 1) valid to invalid --> Call ingress delete
 * 2) invalid to valid --> Call ingress add
 */

func AddNamespaceEventHandler(numWorkers uint32, c *AviController) cache.ResourceEventHandler {
	namespaceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ns := obj.(*corev1.Namespace)
			nsLabels := ns.GetLabels()
			nsFilterObj := utils.GetGlobalNSFilter()
			if utils.CheckIfNamespaceAccepted(ns.GetName(), nsFilterObj, nsLabels, false) {
				utils.AddNamespaceToFilter(ns.GetName(), nsFilterObj)
				utils.AviLog.Debugf("NS Add event: Namespace passed filter: %s", ns.GetName())
			} else {
				//Case: previously deleted valid NS, added back with no labels or invalid labels but nsList contain that ns
				utils.AviLog.Debugf("NS Add event: Namespace did not pass filter: %s", ns.GetName())
				if utils.CheckIfNamespaceAccepted(ns.GetName(), nsFilterObj, nil, true) {
					utils.AviLog.Debugf("Ns Add event: Deleting previous valid namespace: %s from valid NS List", ns.GetName())
					utils.DeleteNamespaceFromFilter(ns.GetName(), nsFilterObj)
				}
			}

		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			nsOld := old.(*corev1.Namespace)
			nsCur := cur.(*corev1.Namespace)
			nsFilterObj := utils.GetGlobalNSFilter()
			if isNamespaceUpdated(nsOld, nsCur) {
				oldNSAccepted := utils.CheckIfNamespaceAccepted(nsOld.GetName(), nsFilterObj, nsOld.Labels, false)
				newNSAccepted := utils.CheckIfNamespaceAccepted(nsCur.GetName(), nsFilterObj, nsCur.Labels, false)

				if !oldNSAccepted && newNSAccepted {
					//Case 1: Namespace updated with valid labels
					//Call ingress add
					utils.AviLog.Debugf("Adding ingresses for namespaces: %s", nsCur.GetName())
					utils.AddNamespaceToFilter(nsCur.GetName(), nsFilterObj)
					if utils.GetInformers().IngressInformer != nil {
						AddIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					} else if utils.GetInformers().RouteInformer != nil {
						AddRoutesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
				} else if oldNSAccepted && !newNSAccepted {
					//Case 2: Old valid namespace updated with invalid labels
					//Call ingress delete
					utils.AviLog.Debugf("Deleting ingresses for namespaces: %s", nsCur.GetName())
					utils.DeleteNamespaceFromFilter(nsCur.GetName(), nsFilterObj)
					if utils.GetInformers().IngressInformer != nil {
						AddIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					} else if utils.GetInformers().RouteInformer != nil {
						AddRoutesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					}
				}

			}

		},
	}
	return namespaceEventHandler
}

func AddRouteEventHandler(numWorkers uint32, c *AviController) cache.ResourceEventHandler {
	routeEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			route := obj.(*routev1.Route)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(route))
			nsFilterObj := utils.GetGlobalNSFilter()
			if !utils.CheckIfNamespaceAccepted(namespace, nsFilterObj, nil, true) {
				utils.AviLog.Debugf("Route add event: Namespace: %s didn't qualify filter. Not adding route", namespace)
				return
			}
			key := utils.OshiftRoute + "/" + utils.ObjKey(route)
			bkt := utils.Bkt(namespace, numWorkers)
			if !lib.HasValidBackends(route.Spec, route.Name, namespace, key) {
				status.UpdateRouteStatusWithErrMsg(key, route.Name, namespace, lib.DuplicateBackends)
			}
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			route, ok := obj.(*routev1.Route)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				route, ok = tombstone.Obj.(*routev1.Route)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Route: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(route))
			nsFilterObj := utils.GetGlobalNSFilter()
			if !utils.CheckIfNamespaceAccepted(namespace, nsFilterObj, nil, true) {
				utils.AviLog.Debugf("Route delete event: Namespace: %s didn't qualify filter. Not deleting route", namespace)
				return
			}
			key := utils.OshiftRoute + "/" + utils.ObjKey(route)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldRoute := old.(*routev1.Route)
			newRoute := cur.(*routev1.Route)
			if isRouteUpdated(oldRoute, newRoute) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(newRoute))
				nsFilterObj := utils.GetGlobalNSFilter()
				if !utils.CheckIfNamespaceAccepted(namespace, nsFilterObj, nil, true) {
					utils.AviLog.Debugf("Route update event: Namespace: %s didn't qualify filter. Not updating route", namespace)
					return
				}
				key := utils.OshiftRoute + "/" + utils.ObjKey(newRoute)
				bkt := utils.Bkt(namespace, numWorkers)
				if !lib.HasValidBackends(newRoute.Spec, newRoute.Name, namespace, key) {
					status.UpdateRouteStatusWithErrMsg(key, newRoute.Name, namespace, lib.DuplicateBackends)
				}
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}
	return routeEventHandler
}

func AddPodEventHandler(numWorkers uint32, c *AviController) cache.ResourceEventHandler {
	podEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			pod := obj.(*corev1.Pod)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(pod))
			key := utils.Pod + "/" + utils.ObjKey(pod)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD\n", key)
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
					utils.AviLog.Errorf("Tombstone contained object that is not an Pod: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(pod))
			key := utils.Pod + "/" + utils.ObjKey(pod)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldPod := old.(*corev1.Pod)
			newPod := cur.(*corev1.Pod)
			if !reflect.DeepEqual(newPod, oldPod) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(newPod))
				key := utils.Pod + "/" + utils.ObjKey(oldPod)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}
	return podEventHandler
}

func (c *AviController) SetupEventHandlers(k8sinfo K8sinformers) {
	cs := k8sinfo.Cs
	utils.AviLog.Debugf("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(utils.AviLog.Debugf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: cs.CoreV1().Events("")})
	mcpQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	c.workqueue = mcpQueue.Workqueue
	numWorkers := mcpQueue.NumWorkers

	epEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			if lib.IsNodePortMode() {
				utils.AviLog.Debugf("skipping endpoint for nodeport mode")
				return
			}
			ep := obj.(*corev1.Endpoints)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ep))
			key := utils.Endpoints + "/" + utils.ObjKey(ep)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			if lib.IsNodePortMode() {
				utils.AviLog.Debugf("skipping endpoint for nodeport mode")
				return
			}
			ep, ok := obj.(*corev1.Endpoints)
			if !ok {
				// endpoints was deleted but its final state is unrecorded.
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
				if lib.IsNodePortMode() {
					utils.AviLog.Debugf("skipping endpoint for nodeport mode")
					return
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(cep))
				key := utils.Endpoints + "/" + utils.ObjKey(cep)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}

	svcEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			svc := obj.(*corev1.Service)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			isSvcLb := isServiceLBType(svc)
			var key string
			if isSvcLb {
				key = utils.L4LBService + "/" + utils.ObjKey(svc)
				if lib.GetAdvancedL4() {
					checkSvcForGatewayPortConflict(svc, key)
				}
				if lib.UseServicesAPI() {
					checkSvcForSvcApiGatewayPortConflict(svc, key)
				}
			} else {
				if lib.GetAdvancedL4() || lib.UseServicesAPI() {
					return
				}
				key = utils.Service + "/" + utils.ObjKey(svc)
			}
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
				// endpoints was deleted but its final state is unrecorded.
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
			isSvcLb := isServiceLBType(svc)
			var key string
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(svc))
			if isSvcLb {
				key = utils.L4LBService + "/" + utils.ObjKey(svc)
			} else {
				if lib.GetAdvancedL4() || lib.UseServicesAPI() {
					return
				}
				key = utils.Service + "/" + utils.ObjKey(svc)
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
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
				isSvcLb := isServiceLBType(svc)
				var key string
				if isSvcLb {
					key = utils.L4LBService + "/" + utils.ObjKey(svc)
					if lib.GetAdvancedL4() {
						checkSvcForGatewayPortConflict(svc, key)
					}
					if lib.UseServicesAPI() {
						checkSvcForSvcApiGatewayPortConflict(svc, key)
					}
				} else {
					if lib.GetAdvancedL4() || lib.UseServicesAPI() {
						return
					}
					key = utils.Service + "/" + utils.ObjKey(svc)
				}

				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			}
		},
	}

	c.informers.EpInformer.Informer().AddEventHandler(epEventHandler)
	c.informers.ServiceInformer.Informer().AddEventHandler(svcEventHandler)

	if lib.GetCNIPlugin() == lib.CALICO_CNI {
		blockAffinityHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				utils.AviLog.Debugf("calico blockaffinity ADD Event")
				if c.DisableSync {
					return
				}
				crd := obj.(*unstructured.Unstructured)
				specJSON, found, err := unstructured.NestedStringMap(crd.UnstructuredContent(), "spec")
				if err != nil || !found {
					utils.AviLog.Warnf("calico blockaffinity spec not found: %+v", err)
					return
				}
				key := utils.NodeObj + "/" + specJSON["name"]
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
			DeleteFunc: func(obj interface{}) {
				utils.AviLog.Debugf("calico blockaffinity DELETE Event")
				if c.DisableSync {
					return
				}
				crd := obj.(*unstructured.Unstructured)
				specJSON, found, err := unstructured.NestedStringMap(crd.UnstructuredContent(), "spec")
				if err != nil || !found {
					utils.AviLog.Warnf("calico blockaffinity spec not found: %+v", err)
					return
				}
				key := utils.NodeObj + "/" + specJSON["name"]
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
		}

		c.dynamicInformers.CalicoBlockAffinityInformer.Informer().AddEventHandler(blockAffinityHandler)
	}

	if lib.GetCNIPlugin() == lib.OPENSHIFT_CNI {
		hostSubnetHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				utils.AviLog.Debugf("hostsubnets ADD Event")
				if c.DisableSync {
					return
				}
				crd := obj.(*unstructured.Unstructured)
				host, found, err := unstructured.NestedString(crd.UnstructuredContent(), "host")
				if err != nil || !found {
					utils.AviLog.Warnf("hostsubnet host not found: %+v", err)
					return
				}

				key := utils.NodeObj + "/" + host
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
			DeleteFunc: func(obj interface{}) {
				utils.AviLog.Debugf("hostsubnets DELETE Event")
				if c.DisableSync {
					return
				}
				crd := obj.(*unstructured.Unstructured)
				host, found, err := unstructured.NestedString(crd.UnstructuredContent(), "host")
				if err != nil || !found {
					utils.AviLog.Warnf("hostsubnet host not found: %+v", err)
					return
				}
				key := utils.NodeObj + "/" + host
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
		}

		c.dynamicInformers.HostSubnetInformer.Informer().AddEventHandler(hostSubnetHandler)
	}

	secretEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			secret := obj.(*corev1.Secret)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
			key := "Secret" + "/" + utils.ObjKey(secret)
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
			if validateAviSecret(secret) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
				key := "Secret" + "/" + utils.ObjKey(secret)
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
				if validateAviSecret(secret) {
					// Only add the key if the resource versions have changed.
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
					key := "Secret" + "/" + utils.ObjKey(secret)
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

	if lib.GetAdvancedL4() {
		// servicesAPI handlers GW/GWClass
		c.SetupAdvL4EventHandlers(numWorkers)
		return
	}

	if lib.UseServicesAPI() {
		c.SetupSvcApiEventHandlers(numWorkers)
	}

	ingressEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ingress := obj.(*networkingv1beta1.Ingress)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			nsFilterObj := utils.GetGlobalNSFilter()
			if utils.CheckIfNamespaceAccepted(namespace, nsFilterObj, nil, true) {
				key := utils.Ingress + "/" + utils.ObjKey(ingress)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
			} else {
				utils.AviLog.Debugf("Ingress add event: Namespace: %s didn't qualify filter. Not adding ingress", namespace)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ingress, ok := obj.(*networkingv1beta1.Ingress)
			if !ok {
				// ingress was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				ingress, ok = tombstone.Obj.(*networkingv1beta1.Ingress)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			nsFilterObj := utils.GetGlobalNSFilter()
			if utils.CheckIfNamespaceAccepted(namespace, nsFilterObj, nil, true) {
				key := utils.Ingress + "/" + utils.ObjKey(ingress)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			} else {
				utils.AviLog.Debugf("Ingress Delete event: NS %s didn't qualify. Not deleting ingress", namespace)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldobj := old.(*networkingv1beta1.Ingress)
			ingress := cur.(*networkingv1beta1.Ingress)
			if isIngressUpdated(oldobj, ingress) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
				nsFilterObj := utils.GetGlobalNSFilter()
				if utils.CheckIfNamespaceAccepted(namespace, nsFilterObj, nil, true) {
					key := utils.Ingress + "/" + utils.ObjKey(ingress)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				} else {
					utils.AviLog.Debugf("Ingress update Event. NS %s didn't qualify. Not updating ingress", namespace)
				}
			}
		},
	}

	nodeEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			node := obj.(*corev1.Node)
			key := utils.NodeObj + "/" + node.Name
			bkt := utils.Bkt(lib.GetTenant(), numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			node, ok := obj.(*corev1.Node)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				node, ok = tombstone.Obj.(*corev1.Node)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Node: %#v", obj)
					return
				}
			}
			key := utils.NodeObj + "/" + node.Name
			bkt := utils.Bkt(lib.GetTenant(), numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldobj := old.(*corev1.Node)
			node := cur.(*corev1.Node)
			key := utils.NodeObj + "/" + node.Name
			if isNodeUpdated(oldobj, node) {
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			} else {
				utils.AviLog.Debugf("key: %s, msg: node object did not change\n", key)
			}
		},
	}

	if c.informers.IngressInformer != nil {
		c.informers.IngressInformer.Informer().AddEventHandler(ingressEventHandler)
	}

	if c.informers.IngressClassInformer != nil {
		ingressClassEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				ingClass := obj.(*networkingv1beta1.IngressClass)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingClass))
				key := utils.IngressClass + "/" + utils.ObjKey(ingClass)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				ingClass, ok := obj.(*networkingv1beta1.IngressClass)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					ingClass, ok = tombstone.Obj.(*networkingv1beta1.IngressClass)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an IngressClass: %#v", obj)
						return
					}
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingClass))
				key := utils.IngressClass + "/" + utils.ObjKey(ingClass)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			},
			UpdateFunc: func(old, cur interface{}) {
				if c.DisableSync {
					return
				}
				oldobj := old.(*networkingv1beta1.IngressClass)
				ingClass := cur.(*networkingv1beta1.IngressClass)
				if oldobj.ResourceVersion != ingClass.ResourceVersion {
					// Only add the key if the resource versions have changed.
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingClass))
					key := utils.IngressClass + "/" + utils.ObjKey(ingClass)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				}
			},
		}

		c.informers.IngressClassInformer.Informer().AddEventHandler(ingressClassEventHandler)
	}

	if lib.GetDisableStaticRoute() && !lib.IsNodePortMode() {
		utils.AviLog.Infof("Static route sync disabled, skipping node informers")
	} else {
		c.informers.NodeInformer.Informer().AddEventHandler(nodeEventHandler)
	}

	if c.informers.RouteInformer != nil {
		routeEventHandler := AddRouteEventHandler(numWorkers, c)
		c.informers.RouteInformer.Informer().AddEventHandler(routeEventHandler)
	}

	// Add CRD handlers HostRule/HTTPRule
	c.SetupAKOCRDEventHandlers(numWorkers)

	//Add namespace event handler if migration is enabled and informer not nil
	nsFilterObj := utils.GetGlobalNSFilter()
	if nsFilterObj.EnableMigration && c.informers.NSInformer != nil {
		utils.AviLog.Debug("Adding namespace event handler")
		namespaceEventHandler := AddNamespaceEventHandler(numWorkers, c)
		c.informers.NSInformer.Informer().AddEventHandler(namespaceEventHandler)
	}

	if lib.GetServiceType() == lib.NodePortLocal {
		podEventHandler := AddPodEventHandler(numWorkers, c)
		c.informers.PodInformer.Informer().AddEventHandler(podEventHandler)
	}
}

func validateAviConfigMap(obj interface{}) (*corev1.ConfigMap, bool) {
	configMap, ok := obj.(*corev1.ConfigMap)
	if ok && lib.GetNamespaceToSync() != "" {
		// AKO is running for a particular namespace, look for the Avi config map here
		if configMap.Name == lib.AviConfigMap {
			return configMap, true
		}
	} else if ok && configMap.Namespace == utils.GetAKONamespace() && configMap.Name == lib.AviConfigMap {
		return configMap, true
	}
	return nil, false
}

func validateAviSecret(secret *corev1.Secret) bool {
	if secret.Namespace == utils.GetAKONamespace() && secret.Name == lib.AviSecret {
		// if the secret is updated or deleted we shutdown API server
		utils.AviLog.Warnf("Avi Secret object %s/%s updated/deleted, shutting down AKO", secret.Namespace, secret.Name)
		lib.ShutdownApi()
		return false
	}
	return true
}

func (c *AviController) Start(stopCh <-chan struct{}) {
	go c.informers.ServiceInformer.Informer().Run(stopCh)
	go c.informers.EpInformer.Informer().Run(stopCh)
	go c.informers.SecretInformer.Informer().Run(stopCh)

	informersList := []cache.InformerSynced{
		c.informers.EpInformer.Informer().HasSynced,
		c.informers.ServiceInformer.Informer().HasSynced,
		c.informers.SecretInformer.Informer().HasSynced,
	}

	if lib.GetServiceType() == lib.NodePortLocal {
		go c.informers.PodInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.PodInformer.Informer().HasSynced)
	}
	if lib.GetCNIPlugin() == lib.CALICO_CNI {
		go c.dynamicInformers.CalicoBlockAffinityInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.dynamicInformers.CalicoBlockAffinityInformer.Informer().HasSynced)
	}
	if lib.GetCNIPlugin() == lib.OPENSHIFT_CNI {
		go c.dynamicInformers.HostSubnetInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.dynamicInformers.HostSubnetInformer.Informer().HasSynced)
	}

	// Disable all informers if we are in advancedL4 mode. We expect to only provide L4 load balancing capability for this feature.
	if lib.GetAdvancedL4() {
		go lib.GetAdvL4Informers().GatewayClassInformer.Informer().Run(stopCh)
		go lib.GetAdvL4Informers().GatewayInformer.Informer().Run(stopCh)

		if !cache.WaitForCacheSync(stopCh, lib.GetAdvL4Informers().GatewayClassInformer.Informer().HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for GatewayClass caches to sync"))
		}
		if !cache.WaitForCacheSync(stopCh, lib.GetAdvL4Informers().GatewayInformer.Informer().HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for Gateway caches to sync"))
		}
		utils.AviLog.Info("Service APIs v1alphapre1 caches synced")
	} else {
		if lib.UseServicesAPI() {
			go lib.GetSvcAPIInformers().GatewayClassInformer.Informer().Run(stopCh)
			go lib.GetSvcAPIInformers().GatewayInformer.Informer().Run(stopCh)

			if !cache.WaitForCacheSync(stopCh, lib.GetSvcAPIInformers().GatewayClassInformer.Informer().HasSynced) {
				runtime.HandleError(fmt.Errorf("Timed out waiting for GatewayClass caches to sync"))
			}
			if !cache.WaitForCacheSync(stopCh, lib.GetSvcAPIInformers().GatewayInformer.Informer().HasSynced) {
				runtime.HandleError(fmt.Errorf("Timed out waiting for Gateway caches to sync"))
			}
			utils.AviLog.Info("Service APIs caches synced")
		}
		if c.informers.IngressInformer != nil {
			go c.informers.IngressInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.IngressInformer.Informer().HasSynced)
		}

		if c.informers.RouteInformer != nil {
			go c.informers.RouteInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.RouteInformer.Informer().HasSynced)
		}

		if c.informers.IngressClassInformer != nil {
			go c.informers.IngressClassInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.IngressClassInformer.Informer().HasSynced)
		}

		go c.informers.NSInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.NSInformer.Informer().HasSynced)

		go c.informers.NodeInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.NodeInformer.Informer().HasSynced)

		go lib.GetCRDInformers().HostRuleInformer.Informer().Run(stopCh)
		go lib.GetCRDInformers().HTTPRuleInformer.Informer().Run(stopCh)
		go lib.GetCRDInformers().NsxAlbInfraSettingInformer.Informer().Run(stopCh)
		if !cache.WaitForCacheSync(stopCh, lib.GetCRDInformers().NsxAlbInfraSettingInformer.Informer().HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for NsxAlbInfraSettingInformer caches to sync"))
		}
		// separate wait steps to try getting hostrules synced first,
		// since httprule has a key relation to hostrules.
		if !cache.WaitForCacheSync(stopCh, lib.GetCRDInformers().HostRuleInformer.Informer().HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for HostRule caches to sync"))
		}
		if !cache.WaitForCacheSync(stopCh, lib.GetCRDInformers().HTTPRuleInformer.Informer().HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for HTTPRule caches to sync"))
		}
		utils.AviLog.Info("CRD caches synced")
	}

	if !cache.WaitForCacheSync(stopCh, informersList...) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Info("Caches synced")
	}
}

func isServiceLBType(svcObj *corev1.Service) bool {
	// If we don't find a service or it is not of type loadbalancer - return false.
	if svcObj.Spec.Type == "LoadBalancer" {
		return true
	}
	return false
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *AviController) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()

	utils.AviLog.Info("Started the Kubernetes Controller")
	<-stopCh
	utils.AviLog.Info("Shutting down the Kubernetes Controller")

	return nil
}

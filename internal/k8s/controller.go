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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	oshiftclient "github.com/openshift/client-go/route/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var controllerInstance *AviController
var ctrlonce sync.Once

// These tags below are only applicable in case of advanced L4 features at the moment.

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;patch;update
// +kubebuilder:rbac:groups=nsx.vmware.com,resources=namespacenetworkinfos;namespacenetworkinfos/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=nsx.vmware.com,resources=clusternetworkinfos;clusternetworkinfos/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=extensions;networking.k8s.io,resources=ingresses;ingresses/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=extensions;networking.k8s.io,resources=ingressclasses;ingressclasses/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gateways;gateways/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=networking.x-k8s.io,resources=gatewayclasses;gatewayclasses/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;delete;update;patch
// +kubebuilder:rbac:groups=core,resources=services;services/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=topology.tanzu.vmware.com,resources=availabilityzones,verbs=get;list;watch
// +kubebuilder:rbac:groups=crd.nsx.vmware.com,resources=vpcnetworkconfigurations,verbs=get;list;watch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=aviinfrasettings;aviinfrasettings/status,verbs=get;list;watch;create;update;patch;delete

type AviController struct {
	worker_id uint32
	//recorder        record.EventRecorder
	informers        *utils.Informers
	dynamicInformers *lib.DynamicInformers
	workqueue        []workqueue.RateLimitingInterface
	DisableSync      bool
	State            *State
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
			State:            &State{},
		}
	})
	return controllerInstance
}

func isNodeUpdated(oldNode, newNode *corev1.Node) bool {
	oldPodCIDRAnnotation, oldOk := oldNode.Annotations[lib.StaticRouteAnnotation]
	newPodCIDRAnnotation, newOk := newNode.Annotations[lib.StaticRouteAnnotation]
	if oldOk != newOk ||
		(oldOk && newOk && oldPodCIDRAnnotation != newPodCIDRAnnotation) {
		return true
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

	if !reflect.DeepEqual(oldNode.Labels, newNode.Labels) {
		return true
	}

	cniPlugin := lib.GetCNIPlugin()
	if (cniPlugin == lib.CALICO_CNI) && (!reflect.DeepEqual(oldNode.Annotations[lib.CalicoIPv4AddressAnnotation], newNode.Annotations[lib.CalicoIPv4AddressAnnotation]) ||
		!reflect.DeepEqual(oldNode.Annotations[lib.CalicoIPv6AddressAnnotation], newNode.Annotations[lib.CalicoIPv6AddressAnnotation])) {
		return true
	}
	if cniPlugin == lib.ANTREA_CNI && !reflect.DeepEqual(oldNode.Annotations[lib.AntreaTransportAddressAnnotation], newNode.Annotations[lib.AntreaTransportAddressAnnotation]) {
		return true
	}

	return false
}

// Consider an ingress has been updated only if spec/annotation is updated
func isIngressUpdated(oldIngress, newIngress *networkingv1.Ingress) bool {
	if oldIngress.ResourceVersion == newIngress.ResourceVersion {
		return false
	}

	if newIngress.GetDeletionTimestamp() != nil {
		return true
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldIngress.Spec))
	newSpecHash := utils.Hash(utils.Stringify(newIngress.Spec))

	// Check for annotation change apart from the ones AKO fills in
	// after status update
	oldAnnotation := oldIngress.DeepCopy().Annotations
	delete(oldAnnotation, lib.VSAnnotation)
	delete(oldAnnotation, lib.ControllerAnnotation)
	newAnnotation := newIngress.DeepCopy().Annotations
	delete(newAnnotation, lib.VSAnnotation)
	delete(newAnnotation, lib.ControllerAnnotation)

	oldAnnotationHash := utils.Hash(utils.Stringify(oldAnnotation))
	newAnnotationHash := utils.Hash(utils.Stringify(newAnnotation))

	if oldSpecHash != newSpecHash || oldAnnotationHash != newAnnotationHash {
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
	oldTenant := oldNS.Annotations[lib.TenantAnnotation]
	newTenant := newNS.Annotations[lib.TenantAnnotation]
	return oldLabelHash != newLabelHash || oldTenant != newTenant
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
		lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
		lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}
}

func AddServicesFromNSToIngestionQueue(numWorkers uint32, c *AviController, namespace string, msg string) {
	var key string
	svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the services during namespace sync: %s", err)
		return
	}
	for _, svcObj := range svcObjs {
		isSvcLb := isServiceLBType(svcObj)
		//Add L4 and Cluster API services to queue
		if isSvcLb && !lib.GetLayer7Only() {
			key = utils.L4LBService + "/" + utils.ObjKey(svcObj)
			if lib.UseServicesAPI() {
				checkSvcForSvcApiGatewayPortConflict(svcObj, key)
			}
		} else {
			key = utils.Service + "/" + utils.ObjKey(svcObj)
		}
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}
}

func AddGatewaysFromNSToIngestionQueue(numWorkers uint32, c *AviController, namespace string, msg string) {
	//TODO: Add code for gateway
	gatewayObjs, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Lister().Gateways(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the gateways during namespace sync: %s", err)
		return
	}
	for _, gatewayObj := range gatewayObjs {
		key := lib.Gateway + "/" + utils.ObjKey(gatewayObj)
		InformerStatusUpdatesForSvcApiGateway(key, gatewayObj)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}
}

func AddMultiClusterIngressFromNSToIngestionQueue(numWorkers uint32, c *AviController, namespace string, msg string) {
	mciObjs, err := utils.GetInformers().MultiClusterIngressInformer.Lister().MultiClusterIngresses(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Error occurred while retrieving multi-cluster ingress for namespace: %s", namespace)
		return
	}
	for _, mciObj := range mciObjs {
		key := lib.MultiClusterIngress + "/" + utils.ObjKey(mciObj)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}
}

func AddServiceImportsFromNSToIngestionQueue(numWorkers uint32, c *AviController, namespace string, msg string) {
	siObjs, err := utils.GetInformers().ServiceImportInformer.Lister().ServiceImports(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Error occurred while retrieving service import for namespace: %s", namespace)
		return
	}
	for _, siObj := range siObjs {
		key := lib.MultiClusterIngress + "/" + utils.ObjKey(siObj)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
/* Namespace update event:  2 cases to handle : NS label changed from 1) valid to invalid --> Call ingress and service delete
 * 2) invalid to valid --> Call ingress and service add
 */
func AddNamespaceEventHandler(numWorkers uint32, c *AviController) cache.ResourceEventHandler {
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

				oldTenant := nsOld.Annotations[lib.TenantAnnotation]
				newTenant := nsCur.Annotations[lib.TenantAnnotation]

				if !oldNSAccepted && newNSAccepted || oldTenant != newTenant {
					//Case 1: Namespace updated with valid labels
					//Call ingress/route and service add
					utils.AddNamespaceToFilter(nsCur.GetName())
					if utils.GetInformers().IngressInformer != nil {
						utils.AviLog.Debugf("Adding ingresses for namespaces: %s", nsCur.GetName())
						AddIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					} else if utils.GetInformers().RouteInformer != nil {
						utils.AviLog.Debugf("Adding routes for namespaces: %s", nsCur.GetName())
						AddRoutesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
					if utils.GetInformers().ServiceInformer != nil {
						utils.AviLog.Debugf("Adding L4 services for namespaces: %s", nsCur.GetName())
						AddServicesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
					if lib.UseServicesAPI() {
						utils.AviLog.Debugf("Adding Gatways for namespaces: %s", nsCur.GetName())
						AddGatewaysFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
					if !oldNSAccepted && newNSAccepted {
						if utils.GetInformers().MultiClusterIngressInformer != nil {
							utils.AviLog.Debugf("Adding multi-cluster ingresses for namespaces: %s", nsCur.GetName())
							AddMultiClusterIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
						}
						if utils.GetInformers().ServiceImportInformer != nil {
							utils.AviLog.Debugf("Adding service imports for namespaces: %s", nsCur.GetName())
							AddServiceImportsFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
						}
					}
				} else if oldNSAccepted && !newNSAccepted {
					//Case 2: Old valid namespace updated with invalid labels
					//Call ingress/route and service delete
					utils.DeleteNamespaceFromFilter(nsCur.GetName())
					if utils.GetInformers().IngressInformer != nil {
						utils.AviLog.Debugf("Deleting ingresses for namespaces: %s", nsCur.GetName())
						AddIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					} else if utils.GetInformers().RouteInformer != nil {
						utils.AviLog.Debugf("Deleting routes for namespaces: %s", nsCur.GetName())
						AddRoutesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					}
					if utils.GetInformers().MultiClusterIngressInformer != nil {
						utils.AviLog.Debugf("Deleting multi-cluster ingress for namespaces: %s", nsCur.GetName())
						AddMultiClusterIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					}
					if utils.GetInformers().ServiceImportInformer != nil {
						utils.AviLog.Debugf("Deleting service imports for namespaces: %s", nsCur.GetName())
						AddServiceImportsFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterDelete)
					}
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

func AddNamespaceAnnotationEventHandler(numWorkers uint32, c *AviController) cache.ResourceEventHandler {
	nsEventHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			nsOld := old.(*corev1.Namespace)
			nsCur := cur.(*corev1.Namespace)
			if isNamespaceUpdated(nsOld, nsCur) {
				oldTenant := nsOld.Annotations[lib.TenantAnnotation]
				newTenant := nsCur.Annotations[lib.TenantAnnotation]
				if oldTenant != newTenant {
					if utils.GetInformers().IngressInformer != nil {
						utils.AviLog.Debugf("Adding ingresses for namespaces: %s", nsCur.GetName())
						AddIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					} else if utils.GetInformers().RouteInformer != nil {
						utils.AviLog.Debugf("Adding routes for namespaces: %s", nsCur.GetName())
						AddRoutesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
					if utils.GetInformers().ServiceInformer != nil {
						utils.AviLog.Debugf("Adding L4 services for namespaces: %s", nsCur.GetName())
						AddServicesFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
					if lib.UseServicesAPI() {
						utils.AviLog.Debugf("Adding Gatways for namespaces: %s", nsCur.GetName())
						AddGatewaysFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
					}
				}
			}
		},
	}
	return nsEventHandler
}

func AddRouteEventHandler(numWorkers uint32, c *AviController) cache.ResourceEventHandler {
	routeEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			route := obj.(*routev1.Route)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(route))
			key := utils.OshiftRoute + "/" + utils.ObjKey(route)
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Route add event: Namespace: %s didn't qualify filter. Not adding route", key, namespace)
				return
			}
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == route.ResourceVersion {
				utils.AviLog.Debugf("key : %s, msg: same resource version returning", key)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			if !lib.HasValidBackends(route.Spec, route.Name, namespace, key) {
				status.UpdateRouteStatusWithErrMsg(key, route.Name, namespace, lib.DuplicateBackends)
			}
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
			key := utils.OshiftRoute + "/" + utils.ObjKey(route)
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Route delete event: Namespace: %s didn't qualify filter. Not deleting route", key, namespace)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			objects.SharedResourceVerInstanceLister().Delete(key)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldRoute := old.(*routev1.Route)
			newRoute := cur.(*routev1.Route)
			if oldRoute.ResourceVersion != newRoute.ResourceVersion || !reflect.DeepEqual(newRoute.Annotations, oldRoute.Annotations) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(newRoute))
				key := utils.OshiftRoute + "/" + utils.ObjKey(newRoute)
				if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Route update event: Namespace: %s didn't qualify filter. Not updating route", key, namespace)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				if !lib.HasValidBackends(newRoute.Spec, newRoute.Name, namespace, key) {
					status.UpdateRouteStatusWithErrMsg(key, newRoute.Name, namespace, lib.DuplicateBackends)
				}
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
					utils.AviLog.Errorf("Tombstone contained object that is not an Pod: %#v", obj)
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
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
		},
	}
	return podEventHandler
}

func (c *AviController) SetupEventHandlers(k8sinfo K8sinformers) {
	mcpQueue := utils.SharedWorkQueue().GetQueueByName(utils.ObjectIngestionLayer)
	c.workqueue = mcpQueue.Workqueue
	numWorkers := mcpQueue.NumWorkers
	var epEventHandler cache.ResourceEventHandlerFuncs
	if lib.GetServiceType() != lib.NodePortLocal {
		epEventHandler = cache.ResourceEventHandlerFuncs{
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
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
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
				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Endpoint Update event: Namespace: %s didn't qualify filter", key, namespace)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				}
			},
		}
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
			if isSvcLb && !lib.GetLayer7Only() {
				//L4 Namespace sync not applicable for advance L4 and service API
				key = utils.L4LBService + "/" + utils.ObjKey(svc)
				if !lib.ValidateSvcforClass(key, svc) && svc.Annotations[lib.SharedVipSvcLBAnnotation] == "" {
					// optimisation for regular LoadBalancer Services
					// not checking for sharedvip as all sharedvip svcs need to be ingested for validation
					return
				}
				if lib.IsNamespaceBlocked(namespace) || !utils.IsServiceNSValid(namespace) {
					utils.AviLog.Debugf("key: %s, msg: L4 Service add event: Namespace: %s didn't qualify filter. Not adding service.", key, namespace)
					return
				}
				if utils.IsWCP() {
					checkSvcForGatewayPortConflict(svc, key)
				}
				if svc.Annotations[lib.SharedVipSvcLBAnnotation] != "" {
					// mark the object type as ShareVipSvc
					// to separate these out from regulare clusterip, svclb services
					key = lib.SharedVipServiceKey + "/" + utils.ObjKey(svc)
				}
			} else {
				if lib.IsNamespaceBlocked(namespace) || utils.GetAdvancedL4() || !utils.CheckIfNamespaceAccepted(namespace) {
					return
				}
				if lib.UseServicesAPI() {
					checkSvcForSvcApiGatewayPortConflict(svc, key)
				}
				key = utils.Service + "/" + utils.ObjKey(svc)
			}
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == svc.ResourceVersion {
				utils.AviLog.Debugf("key : %s, msg: same resource version returning", key)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
			if isSvcLb && !lib.GetLayer7Only() {
				key = utils.L4LBService + "/" + utils.ObjKey(svc)
				if !lib.ValidateSvcforClass(key, svc) && svc.Annotations[lib.SharedVipSvcLBAnnotation] == "" {
					return
				}
				if lib.IsNamespaceBlocked(namespace) || !utils.IsServiceNSValid(namespace) {
					utils.AviLog.Debugf("key: %s, msg: L4 Service delete event: Namespace: %s didn't qualify filter. Not deleting service.", key, namespace)
					return
				}

				if svc.Annotations[lib.SharedVipSvcLBAnnotation] != "" {
					// mark the object type as ShareVipSvc
					// to separate these out from regulare clusterip, svclb services
					key = lib.SharedVipServiceKey + "/" + utils.ObjKey(svc)
				}
			} else {
				if lib.IsNamespaceBlocked(namespace) || utils.GetAdvancedL4() || !utils.CheckIfNamespaceAccepted(namespace) {
					return
				}
				key = utils.Service + "/" + utils.ObjKey(svc)
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
				isSvcLb := isServiceLBType(svc)
				var key string
				if isSvcLb && !lib.GetLayer7Only() {
					key = utils.L4LBService + "/" + utils.ObjKey(svc)
					if !lib.ValidateSvcforClass(key, svc) && svc.Annotations[lib.SharedVipSvcLBAnnotation] == "" {
						return
					}
					if lib.IsNamespaceBlocked(namespace) || !utils.IsServiceNSValid(namespace) {
						utils.AviLog.Debugf("key: %s, msg: L4 Service update event: Namespace: %s didn't qualify filter. Not updating service.", key, namespace)
						return
					}
					if utils.IsWCP() {
						checkSvcForGatewayPortConflict(svc, key)
					}
					if svc.Annotations[lib.SharedVipSvcLBAnnotation] != "" {
						key = lib.SharedVipServiceKey + "/" + utils.ObjKey(svc)
					}
				} else {
					if lib.IsNamespaceBlocked(namespace) || utils.GetAdvancedL4() || !utils.CheckIfNamespaceAccepted(namespace) {
						return
					}
					if lib.UseServicesAPI() {
						checkSvcForSvcApiGatewayPortConflict(svc, key)
					}
					key = utils.Service + "/" + utils.ObjKey(svc)
				}

				bkt := utils.Bkt(namespace, numWorkers)
				var oldKey string
				if isSvcLb &&
					!lib.GetLayer7Only() &&
					oldobj.Annotations[lib.SharedVipSvcLBAnnotation] != svc.Annotations[lib.SharedVipSvcLBAnnotation] {
					if oldobj.Annotations[lib.SharedVipSvcLBAnnotation] != "" {
						// Handles annotation -> no-annotation transition, old pool needs to be deleted.
						oldKey = lib.SharedVipServiceKey + "/" + utils.ObjKey(oldobj)
					} else {
						// Handles no-annotation -> annotation transition too, old L4 VS is deleted.
						oldKey = utils.L4LBService + "/" + utils.ObjKey(oldobj)
					}
				}
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				if oldKey != "" && key != oldKey {
					c.workqueue[bkt].AddRateLimited(oldKey)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", oldKey)
				}
			}
		},
	}

	if lib.GetServiceType() != lib.NodePortLocal {
		c.informers.EpInformer.Informer().AddEventHandler(epEventHandler)
	}
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
				key := utils.NodeObj + "/" + specJSON["node"]
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
				key := utils.NodeObj + "/" + specJSON["node"]
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
		}

		c.dynamicInformers.HostSubnetInformer.Informer().AddEventHandler(hostSubnetHandler)
	}

	if lib.GetCNIPlugin() == lib.CILIUM_CNI {
		ciliumNodeHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				utils.AviLog.Debugf("ciliumnode ADD Event")
				if c.DisableSync {
					return
				}
				crd := obj.(*unstructured.Unstructured)
				nodename := crd.GetName()
				key := utils.NodeObj + "/" + nodename
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				utils.AviLog.Debugf("ciliumnode UPDATE Event")
				if c.DisableSync {
					return
				}
				crd := newObj.(*unstructured.Unstructured)
				nodename := crd.GetName()
				key := utils.NodeObj + "/" + nodename
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
			DeleteFunc: func(obj interface{}) {
				utils.AviLog.Debugf("ciliumnode DELETE Event")
				if c.DisableSync {
					return
				}
				crd := obj.(*unstructured.Unstructured)
				nodename := crd.GetName()
				key := utils.NodeObj + "/" + nodename
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
		}

		c.dynamicInformers.CiliumNodeInformer.Informer().AddEventHandler(ciliumNodeHandler)
	}

	secretEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			secret := obj.(*corev1.Secret)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(secret))
			key := "Secret" + "/" + utils.ObjKey(secret)
			if lib.IsIstioEnabled() && secret.Namespace == utils.GetAKONamespace() && secret.Name == lib.IstioSecret {
				key = utils.Secret + "/" + utils.GetAKONamespace() + "/" + lib.IstioSecret
				utils.AviLog.Infof("key: %s, msg: Istio Secret ADD", key)
			}
			if lib.IsNamespaceBlocked(namespace) {
				utils.AviLog.Debugf("key: %s, msg: secret add event. namespace: %s didn't qualify filter", key, namespace)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			secret, ok := obj.(*corev1.Secret)
			if lib.IsIstioEnabled() && secret.Namespace == utils.GetAKONamespace() && secret.Name == lib.IstioSecret {
				utils.AviLog.Warnf("Istio secret deleted")
			}
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
				key := "Secret" + "/" + utils.ObjKey(secret)
				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key: %s, msg: secret delete event. namespace: %s didn't qualify filter", key, namespace)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
					key := "Secret" + "/" + utils.ObjKey(secret)
					if lib.IsIstioEnabled() && secret.Namespace == utils.GetAKONamespace() && secret.Name == lib.IstioSecret {
						key = utils.Secret + "/" + utils.GetAKONamespace() + "/" + lib.IstioSecret
						utils.AviLog.Infof("key: %s, msg: Istio Secret UPDATE", key)
					}
					if lib.IsNamespaceBlocked(namespace) {
						utils.AviLog.Debugf("key: %s, msg: secret update event. namespace: %s didn't qualify filter", key, namespace)
						return
					}
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				}
			}
		},
	}

	if c.informers.SecretInformer != nil {
		c.informers.SecretInformer.Informer().AddEventHandler(secretEventHandler)
	}

	// Add CRD handlers HostRule/HTTPRule/AviInfraSettings/SSORule
	c.SetupAKOCRDEventHandlers(numWorkers)

	if c.informers.NSInformer != nil {
		nsEventHandler := AddNamespaceAnnotationEventHandler(numWorkers, c)
		c.informers.NSInformer.Informer().AddEventHandler(nsEventHandler)
	}

	if utils.IsWCP() {
		// servicesAPI handlers GW/GWClass
		c.SetupAdvL4EventHandlers(numWorkers)
		c.SetupNamespaceEventHandler(numWorkers)
		if utils.GetAdvancedL4() {
			return
		}
	}

	if lib.UseServicesAPI() {
		c.SetupSvcApiEventHandlers(numWorkers)
	}

	ingressEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ingress := obj.(*networkingv1.Ingress)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			key := utils.Ingress + "/" + utils.ObjKey(ingress)
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Ingress add event: Namespace: %s didn't qualify filter. Not adding ingress", key, namespace)
				return
			}

			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == ingress.ResourceVersion {
				utils.AviLog.Debugf("key : %s, msg: same resource version returning", key)
				return
			}
			// optimization to check if ingress belongs to ingressClass handled by AKO.
			if !lib.ValidateIngressForClass(key, ingress) {
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			ingress, ok := obj.(*networkingv1.Ingress)
			if !ok {
				// ingress was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				ingress, ok = tombstone.Obj.(*networkingv1.Ingress)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
					return
				}
			}
			key := utils.Ingress + "/" + utils.ObjKey(ingress)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Ingress Delete event: Namespace: %s didn't qualify filter. Not deleting ingress", key, namespace)
				return
			}

			// optimization to check if ingress belongs to ingressClass handled by AKO.
			if !lib.ValidateIngressForClass(key, ingress) {
				return
			}
			objects.SharedResourceVerInstanceLister().Delete(key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldobj := old.(*networkingv1.Ingress)
			ingress := cur.(*networkingv1.Ingress)
			if isIngressUpdated(oldobj, ingress) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingress))
				key := utils.Ingress + "/" + utils.ObjKey(ingress)
				if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Ingress Update event: Namespace: %s didn't qualify filter. Not updating ingress", key, namespace)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
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
			if lib.IsNodePortMode() {
				if !lib.IsValidLabelOnNode(node.GetLabels(), key) {
					utils.AviLog.Debugf("key: %s, msg: no valid labels on a node", key)
					return
				}
			}

			bkt := utils.Bkt(lib.GetTenant(), numWorkers)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == node.ResourceVersion {
				utils.AviLog.Debugf("key : %s, msg: same resource version returning", key)
				return
			}
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
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
			if lib.IsNodePortMode() {
				if !lib.IsValidLabelOnNode(node.GetLabels(), key) {
					utils.AviLog.Debugf("key: %s, msg: no valid labels on a node", key)
					return
				}
			}
			bkt := utils.Bkt(lib.GetTenant(), numWorkers)
			objects.SharedResourceVerInstanceLister().Delete(key)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			oldobj := old.(*corev1.Node)
			node := cur.(*corev1.Node)
			key := utils.NodeObj + "/" + node.Name
			if lib.IsNodePortMode() {
				if !lib.IsValidLabelOnNode(node.GetLabels(), key) && !lib.IsValidLabelOnNode(oldobj.GetLabels(), key) {
					utils.AviLog.Debugf("key: %s, msg: no valid labels on a node", key)
					return
				}
			}
			if isNodeUpdated(oldobj, node) {
				bkt := utils.Bkt(lib.GetTenant(), numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
			} else {
				utils.AviLog.Debugf("key: %s, msg: node object did not change", key)
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
				ingClass := obj.(*networkingv1.IngressClass)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingClass))
				key := utils.IngressClass + "/" + utils.ObjKey(ingClass)
				ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
				if ok && resVer.(string) == ingClass.ResourceVersion {
					utils.AviLog.Debugf("key : %s, msg: same resource version returning", key)
					return
				}
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				ingClass, ok := obj.(*networkingv1.IngressClass)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					ingClass, ok = tombstone.Obj.(*networkingv1.IngressClass)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an IngressClass: %#v", obj)
						return
					}
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingClass))
				key := utils.IngressClass + "/" + utils.ObjKey(ingClass)
				bkt := utils.Bkt(namespace, numWorkers)
				objects.SharedResourceVerInstanceLister().Delete(key)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			},
			UpdateFunc: func(old, cur interface{}) {
				if c.DisableSync {
					return
				}
				oldobj := old.(*networkingv1.IngressClass)
				ingClass := cur.(*networkingv1.IngressClass)
				if oldobj.ResourceVersion != ingClass.ResourceVersion {
					// Only add the key if the resource versions have changed.
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ingClass))
					key := utils.IngressClass + "/" + utils.ObjKey(ingClass)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				}
			},
		}

		c.informers.IngressClassInformer.Informer().AddEventHandler(ingressClassEventHandler)

	}

	if lib.GetDisableStaticRoute() && !lib.IsNodePortMode() {
		utils.AviLog.Infof("Static route sync disabled, skipping node informers")
	} else {
		//For all AKO, node event handler can be enabled.
		c.informers.NodeInformer.Informer().AddEventHandler(nodeEventHandler)
	}

	if c.informers.RouteInformer != nil {
		routeEventHandler := AddRouteEventHandler(numWorkers, c)
		c.informers.RouteInformer.Informer().AddEventHandler(routeEventHandler)
	}

	// Add MultiClusterIngress and ServiceImport CRD event handlers
	if utils.IsMultiClusterIngressEnabled() {
		c.SetupMultiClusterIngressEventHandlers(numWorkers)
		c.SetupServiceImportEventHandlers(numWorkers)
	}

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

func checkAviSecretUpdateAndShutdown(secret *corev1.Secret) bool {
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
	if lib.GetCNIPlugin() == lib.CALICO_CNI {
		go c.dynamicInformers.CalicoBlockAffinityInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.dynamicInformers.CalicoBlockAffinityInformer.Informer().HasSynced)
	}
	if lib.GetCNIPlugin() == lib.OPENSHIFT_CNI {
		go c.dynamicInformers.HostSubnetInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.dynamicInformers.HostSubnetInformer.Informer().HasSynced)
	}
	if lib.GetCNIPlugin() == lib.CILIUM_CNI {
		go c.dynamicInformers.CiliumNodeInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.dynamicInformers.CiliumNodeInformer.Informer().HasSynced)
	}

	if utils.IsVCFCluster() {
		if c.informers.IngressClassInformer != nil {
			go c.informers.IngressClassInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.IngressClassInformer.Informer().HasSynced)
		}

		if c.informers.IngressInformer != nil {
			go c.informers.IngressInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.IngressInformer.Informer().HasSynced)
		}
	}

	// Disable all informers if we are in advancedL4 mode. We expect to only provide L4 load balancing capability for this feature.
	if utils.IsWCP() {
		go lib.AKOControlConfig().AdvL4Informers().GatewayClassInformer.Informer().Run(stopCh)
		informersList = append(informersList, lib.AKOControlConfig().AdvL4Informers().GatewayClassInformer.Informer().HasSynced)
		go lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Informer().Run(stopCh)
		informersList = append(informersList, lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Informer().HasSynced)
		go lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Informer().Run(stopCh)
		informersList = append(informersList, lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Informer().HasSynced)
	} else {
		if lib.UseServicesAPI() {
			go lib.AKOControlConfig().SvcAPIInformers().GatewayClassInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().SvcAPIInformers().GatewayClassInformer.Informer().HasSynced)
			go lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Informer().HasSynced)
		}

		if c.informers.IngressClassInformer != nil {
			go c.informers.IngressClassInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.IngressClassInformer.Informer().HasSynced)
		}

		if c.informers.IngressInformer != nil {
			go c.informers.IngressInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.IngressInformer.Informer().HasSynced)
		}

		if c.informers.RouteInformer != nil {
			go c.informers.RouteInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.RouteInformer.Informer().HasSynced)
		}

		go c.informers.NodeInformer.Informer().Run(stopCh)
		informersList = append(informersList, c.informers.NodeInformer.Informer().HasSynced)

		if lib.AKOControlConfig().AviInfraSettingEnabled() {
			go lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Informer().HasSynced)

		}

		// separate wait steps to try getting hostrules synced first,
		// since httprule has a key relation to hostrules.

		if lib.AKOControlConfig().L7RuleEnabled() {
			go lib.AKOControlConfig().CRDInformers().L7RuleInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().CRDInformers().L7RuleInformer.Informer().HasSynced)
		}

		if lib.AKOControlConfig().HostRuleEnabled() {
			go lib.AKOControlConfig().CRDInformers().HostRuleInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().CRDInformers().HostRuleInformer.Informer().HasSynced)
		}

		if lib.AKOControlConfig().HttpRuleEnabled() {
			go lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Informer().HasSynced)
		}

		if lib.AKOControlConfig().SsoRuleEnabled() {
			go lib.AKOControlConfig().CRDInformers().SSORuleInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().CRDInformers().SSORuleInformer.Informer().HasSynced)
		}

		if lib.AKOControlConfig().L4RuleEnabled() {
			go lib.AKOControlConfig().CRDInformers().L4RuleInformer.Informer().Run(stopCh)
			informersList = append(informersList, lib.AKOControlConfig().CRDInformers().L4RuleInformer.Informer().HasSynced)
		}

		if utils.IsMultiClusterIngressEnabled() {
			go c.informers.MultiClusterIngressInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.MultiClusterIngressInformer.Informer().HasSynced)
			go c.informers.ServiceImportInformer.Informer().Run(stopCh)
			informersList = append(informersList, c.informers.ServiceImportInformer.Informer().HasSynced)
		}
	}

	if !cache.WaitForCacheSync(stopCh, informersList...) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced")
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

	utils.AviLog.Infof("Started the Kubernetes Controller")
	<-stopCh
	utils.AviLog.Infof("Shutting down the Kubernetes Controller")

	return nil
}

func (c *AviController) GetValidator() Validator {
	return NewValidator()
}

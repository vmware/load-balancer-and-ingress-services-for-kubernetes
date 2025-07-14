/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	"strconv"
	"strings"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	advl4crd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned"
	advl4informer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/informers/externalversions"
)

func NewAdvL4Informers(cs advl4crd.Interface) {
	var advl4InformerFactory advl4informer.SharedInformerFactory

	advl4InformerFactory = advl4informer.NewSharedInformerFactoryWithOptions(cs, time.Second*30)
	gatewayInformer := advl4InformerFactory.Networking().V1alpha1pre1().Gateways()
	gatewayClassInformer := advl4InformerFactory.Networking().V1alpha1pre1().GatewayClasses()

	lib.AKOControlConfig().SetAdvL4Informers(&lib.AdvL4Informers{
		GatewayInformer:      gatewayInformer,
		GatewayClassInformer: gatewayClassInformer,
	})
}

// SetupAdvL4EventHandlers handles setting up of AdvL4 event handlers
func (c *AviController) SetupAdvL4EventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AdvL4 Event handlers")
	informer := lib.AKOControlConfig().AdvL4Informers()

	gatewayEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gw := obj.(*advl4v1alpha1pre1.Gateway)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
			key := lib.Gateway + "/" + utils.ObjKey(gw)
			if lib.IsNamespaceBlocked(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Gateway add event: namespace %s didn't qualify filter.", key, namespace)
				return
			}
			utils.AviLog.Infof("key: %s, msg: ADD", key)

			InformerStatusUpdatesForGateway(key, gw)
			checkGWForGatewayPortConflict(key, gw)

			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*advl4v1alpha1pre1.Gateway)
			gw := new.(*advl4v1alpha1pre1.Gateway)
			oldAnnotVal := oldObj.Annotations[lib.GwProxyProtocolEnableAnnotation]
			newAnnotVal := gw.Annotations[lib.GwProxyProtocolEnableAnnotation]

			if !reflect.DeepEqual(oldObj.Spec, gw.Spec) || gw.GetDeletionTimestamp() != nil || oldAnnotVal != newAnnotVal {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
				key := lib.Gateway + "/" + utils.ObjKey(gw)
				utils.AviLog.Infof("key: %s, msg: UPDATE", key)
				if lib.IsNamespaceBlocked(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Gateway update event: namespace %s didn't qualify filter.", key, namespace)
					return
				}
				InformerStatusUpdatesForGateway(key, gw)
				checkGWForGatewayPortConflict(key, gw)

				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gw, ok := obj.(*advl4v1alpha1pre1.Gateway)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				gw, ok = tombstone.Obj.(*advl4v1alpha1pre1.Gateway)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an Gateway: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gw))
			key := lib.Gateway + "/" + utils.ObjKey(gw)
			if lib.IsNamespaceBlocked(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Gateway delete event: namespace %s didn't qualify filter.", key, namespace)
				return
			}
			utils.AviLog.Infof("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
	}

	gatewayClassEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gwclass := obj.(*advl4v1alpha1pre1.GatewayClass)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwclass))
			key := lib.GatewayClass + "/" + utils.ObjKey(gwclass)
			utils.AviLog.Infof("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*advl4v1alpha1pre1.GatewayClass)
			gwclass := new.(*advl4v1alpha1pre1.GatewayClass)
			if !reflect.DeepEqual(oldObj.Spec, gwclass.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwclass))
				key := lib.GatewayClass + "/" + utils.ObjKey(gwclass)
				utils.AviLog.Infof("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			gwclass, ok := obj.(*advl4v1alpha1pre1.GatewayClass)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				gwclass, ok = tombstone.Obj.(*advl4v1alpha1pre1.GatewayClass)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an GatewayClass: %#v", obj)
					return
				}
			}
			key := lib.GatewayClass + "/" + utils.ObjKey(gwclass)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(gwclass))
			utils.AviLog.Infof("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
	}

	informer.GatewayInformer.Informer().AddEventHandler(gatewayEventHandler)
	informer.GatewayClassInformer.Informer().AddEventHandler(gatewayClassEventHandler)
}

func (c *AviController) SetupNamespaceEventHandler(numWorkers uint32) {
	nsHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			nsOld := old.(*corev1.Namespace)
			nsCur := cur.(*corev1.Namespace)
			oldTenant := nsOld.Annotations[lib.TenantAnnotation]
			newTenant := nsCur.Annotations[lib.TenantAnnotation]
			oldInfraSetting := nsOld.Annotations[lib.InfraSettingNameAnnotation]
			newInfraSetting := nsCur.Annotations[lib.InfraSettingNameAnnotation]
			if oldTenant != newTenant || oldInfraSetting != newInfraSetting {
				if utils.GetInformers().IngressInformer != nil {
					utils.AviLog.Debugf("Adding ingresses for namespaces: %s", nsCur.GetName())
					AddIngressFromNSToIngestionQueue(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
				}
				utils.AviLog.Debugf("Adding Gateways for namespaces: %s", nsCur.GetName())
				AddGatewaysFromNSToIngestionQueueWCP(numWorkers, c, nsCur.GetName(), lib.NsFilterAdd)
			}
			objKey := utils.ObjKey(nsCur)
			if objKey == "" {
				return
			}
			key := lib.Namespace + "/" + objKey
			utils.AviLog.Infof("key: %s, msg: UPDATE", key)
			bkt := utils.Bkt(nsCur.GetName(), numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		DeleteFunc: func(obj interface{}) {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				ns, ok = tombstone.Obj.(*corev1.Namespace)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a Namespace: %#v", obj)
					return
				}
			}
			objKey := utils.ObjKey(ns)
			if objKey == "" {
				return
			}
			key := lib.Namespace + "/" + objKey
			utils.AviLog.Infof("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(ns.Name, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
	}
	c.informers.NSInformer.Informer().AddEventHandler(nsHandler)
}

func InformerStatusUpdatesForGateway(key string, gateway *advl4v1alpha1pre1.Gateway) {
	gwStatus := gateway.Status.DeepCopy()
	defer status.UpdateGatewayStatusObject(key, gateway, gwStatus)

	status.InitializeGatewayConditions(gwStatus, &gateway.Spec, false)
	gwClassObj, err := lib.AKOControlConfig().AdvL4Informers().GatewayClassInformer.Lister().Get(gateway.Spec.Class)
	if err != nil {
		status.UpdateGatewayStatusGWCondition(key, gwStatus, &status.UpdateGWStatusConditionOptions{
			Type:    "Pending",
			Status:  corev1.ConditionTrue,
			Message: fmt.Sprintf("Corresponding networking.x-k8s.io/gatewayclass not found %s", gateway.Spec.Class),
			Reason:  "InvalidGatewayClass",
		})
		utils.AviLog.Warnf("key: %s, msg: Corresponding networking.x-k8s.io/gatewayclass not found %s %v",
			key, gateway.Spec.Class, err)
		return
	}

	for _, listener := range gateway.Spec.Listeners {
		gwName, nameOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNameLabelKey]
		gwNamespace, nsOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNamespaceLabelKey]
		if !nameOk || !nsOk ||
			(nameOk && gwName != gateway.Name) ||
			(nsOk && gwNamespace != gateway.Namespace) {
			status.UpdateGatewayStatusGWCondition(key, gwStatus, &status.UpdateGWStatusConditionOptions{
				Type:    "Pending",
				Status:  corev1.ConditionTrue,
				Message: "Incorrect gateway matchLabels configuration",
				Reason:  "InvalidMatchLabels",
			})
			return
		}
	}

	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if gwClassObj.Spec.Controller != lib.AviGatewayController {
		// Return an error since this is not our object.
		status.UpdateGatewayStatusGWCondition(key, gwStatus, &status.UpdateGWStatusConditionOptions{
			Type:    "Pending",
			Status:  corev1.ConditionTrue,
			Message: fmt.Sprintf("Unable to identify controller %s", gwClassObj.Spec.Controller),
			Reason:  "UnidentifiedController",
		})
	}
}

func checkSvcForGatewayPortConflict(svc *corev1.Service, key string) {
	gateway, portProtocols := nodes.ParseL4ServiceForGateway(svc, key)
	if gateway == "" {
		utils.AviLog.Warnf("key: %s, msg: Unable to find gateway labels in service", key)
		return
	}

	found, gwSvcListeners := objects.ServiceGWLister().GetGwToSvcs(gateway)
	if !found {
		return
	}

	gwNSName := strings.Split(gateway, "/")
	gw, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(gwNSName[0]).Get(gwNSName[1])
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to find gateway: %v", key, err)
		return
	}

	// detect port conflict
	for _, portProtocol := range portProtocols {
		if val, ok := gwSvcListeners[portProtocol]; ok {
			if !utils.HasElem(val, svc.Namespace+"/"+svc.Name) {
				val = append(val, svc.Namespace+"/"+svc.Name)
			}
			if len(val) > 1 {
				portProtocolArr := strings.Split(portProtocol, "/")
				gwStatus := gw.Status.DeepCopy()
				status.UpdateGatewayStatusListenerConditions(key, gwStatus, portProtocolArr[1], &status.UpdateGWStatusConditionOptions{
					Type:   "PortConflict",
					Status: corev1.ConditionTrue,
					Reason: fmt.Sprintf("conflicting port configuration provided in service %s and %s/%s", val, svc.Namespace, svc.Name),
				})
				status.UpdateGatewayStatusObject(key, gw, gwStatus)
				return
			}
		}
	}

	// detect unsupported protocol
	// TODO
}

func checkGWForGatewayPortConflict(key string, gw *advl4v1alpha1pre1.Gateway) {
	found, gwSvcListeners := objects.ServiceGWLister().GetGwToSvcs(gw.Namespace + "/" + gw.Name)
	if !found {
		return
	}

	var gwProtocols []string
	// port conflicts
	for _, listener := range gw.Spec.Listeners {
		portProtoGW := string(listener.Protocol) + "/" + strconv.Itoa(int(listener.Port))
		if !utils.HasElem(gwProtocols, string(listener.Protocol)) {
			gwProtocols = append(gwProtocols, string(listener.Protocol))
		}

		if val, ok := gwSvcListeners[portProtoGW]; ok && len(val) > 1 {
			gwStatus := gw.Status.DeepCopy()
			status.UpdateGatewayStatusListenerConditions(key, gwStatus, strconv.Itoa(int(listener.Port)), &status.UpdateGWStatusConditionOptions{
				Type:   "PortConflict",
				Status: corev1.ConditionTrue,
				Reason: fmt.Sprintf("conflicting port configuration provided in service %s and %v", val, gwSvcListeners[portProtoGW]),
			})
			status.UpdateGatewayStatusObject(key, gw, gwStatus)
			return
		}
	}

	// unsupported protocol
	for portProto, svcs := range gwSvcListeners {
		svcProtocol := strings.Split(portProto, "/")[0]
		if !utils.HasElem(gwProtocols, svcProtocol) {
			gwStatus := gw.Status.DeepCopy()
			status.UpdateGatewayStatusListenerConditions(key, gwStatus, strings.Split(portProto, "/")[1], &status.UpdateGWStatusConditionOptions{
				Type:   "UnsupportedProtocol",
				Status: corev1.ConditionTrue,
				Reason: fmt.Sprintf("Unsupported protocol found in services %v", svcs),
			})
			status.UpdateGatewayStatusObject(key, gw, gwStatus)
			return
		}
	}
}

func AddGatewaysFromNSToIngestionQueueWCP(numWorkers uint32, c *AviController, namespace string, msg string) {
	gateways, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Warnf("Failed to list Gateways in the namespace %s", namespace)
		return
	}
	for _, gw := range gateways {
		key := lib.Gateway + "/" + utils.ObjKey(gw)
		bkt := utils.Bkt(namespace, numWorkers)
		c.workqueue[bkt].AddRateLimited(key)
		utils.AviLog.Debugf("key: %s, msg: %s for namespace: %s", key, msg, namespace)
	}
}

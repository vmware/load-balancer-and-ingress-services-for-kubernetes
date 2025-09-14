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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"

	istiocrd "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	v1alpha2akoinformers "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/informers/externalversions"
	v1beta1akoinformers "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/informers/externalversions"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

func NewCRDInformers() {
	v1alpha2akoInformerFactory := v1alpha2akoinformers.NewSharedInformerFactoryWithOptions(
		lib.AKOControlConfig().V1alpha2CRDClientset(), time.Second*30)
	ssoRuleInformer := v1alpha2akoInformerFactory.Ako().V1alpha2().SSORules()
	l4RuleInformer := v1alpha2akoInformerFactory.Ako().V1alpha2().L4Rules()
	l7RuleInformer := v1alpha2akoInformerFactory.Ako().V1alpha2().L7Rules()

	//v1beta1 informer initialization
	v1beta1akoInformerFactory := v1beta1akoinformers.NewSharedInformerFactoryWithOptions(
		lib.AKOControlConfig().V1beta1CRDClientset(), time.Second*30)
	aviInfraSettingInformer := v1beta1akoInformerFactory.Ako().V1beta1().AviInfraSettings()
	hostRuleInformer := v1beta1akoInformerFactory.Ako().V1beta1().HostRules()
	httpRuleInformer := v1beta1akoInformerFactory.Ako().V1beta1().HTTPRules()

	lib.AKOControlConfig().SetCRDInformers(&lib.AKOCrdInformers{
		HostRuleInformer:        hostRuleInformer,
		HTTPRuleInformer:        httpRuleInformer,
		SSORuleInformer:         ssoRuleInformer,
		L4RuleInformer:          l4RuleInformer,
		L7RuleInformer:          l7RuleInformer,
		AviInfraSettingInformer: aviInfraSettingInformer,
	})
}

func NewInfraSettingCRDInformer() {
	akoInformerFactory := v1beta1akoinformers.NewSharedInformerFactoryWithOptions(lib.AKOControlConfig().V1beta1CRDClientset(), time.Second*30)
	aviSettingsInformer := akoInformerFactory.Ako().V1beta1().AviInfraSettings()
	lib.AKOControlConfig().SetCRDInformers(&lib.AKOCrdInformers{
		AviInfraSettingInformer: aviSettingsInformer,
	})
}

func NewIstioCRDInformers(cs istiocrd.Interface) {
	var istioInformerFactory istioinformers.SharedInformerFactory

	istioInformerFactory = istioinformers.NewSharedInformerFactoryWithOptions(cs, time.Second*30)
	vsInformer := istioInformerFactory.Networking().V1alpha3().VirtualServices()
	drInformer := istioInformerFactory.Networking().V1alpha3().DestinationRules()
	gatewayInformer := istioInformerFactory.Networking().V1alpha3().Gateways()

	lib.AKOControlConfig().SetIstioCRDInformers(&lib.IstioCRDInformers{
		VirtualServiceInformer:  vsInformer,
		DestinationRuleInformer: drInformer,
		GatewayInformer:         gatewayInformer,
	})
}

func isHostRuleUpdated(oldHostRule, newHostRule *akov1beta1.HostRule) bool {
	if oldHostRule.ResourceVersion == newHostRule.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldHostRule.Spec) + oldHostRule.Status.Status)
	newSpecHash := utils.Hash(utils.Stringify(newHostRule.Spec) + newHostRule.Status.Status)

	return oldSpecHash != newSpecHash
}

func isHTTPRuleUpdated(oldHTTPRule, newHTTPRule *akov1beta1.HTTPRule) bool {
	if oldHTTPRule.ResourceVersion == newHTTPRule.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldHTTPRule.Spec) + oldHTTPRule.Status.Status)
	newSpecHash := utils.Hash(utils.Stringify(newHTTPRule.Spec) + newHTTPRule.Status.Status)

	return oldSpecHash != newSpecHash
}

func isAviInfraUpdated(oldAviInfra, newAviInfra *akov1beta1.AviInfraSetting) bool {
	if oldAviInfra.ResourceVersion == newAviInfra.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldAviInfra.Spec) + oldAviInfra.Status.Status)
	newSpecHash := utils.Hash(utils.Stringify(newAviInfra.Spec) + newAviInfra.Status.Status)

	return oldSpecHash != newSpecHash
}

func isSSORuleUpdated(oldSSORule, newSSORule *akov1alpha2.SSORule) bool {
	if oldSSORule.ResourceVersion == newSSORule.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldSSORule.Spec) + oldSSORule.Status.Status)
	newSpecHash := utils.Hash(utils.Stringify(newSSORule.Spec) + newSSORule.Status.Status)

	return oldSpecHash != newSpecHash
}

func isL4RuleUpdated(oldL4Rule, newL4Rule *akov1alpha2.L4Rule) bool {
	if oldL4Rule.ResourceVersion == newL4Rule.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldL4Rule.Spec) + oldL4Rule.Status.Status)
	newSpecHash := utils.Hash(utils.Stringify(newL4Rule.Spec) + newL4Rule.Status.Status)

	return oldSpecHash != newSpecHash
}

func isL7RuleUpdated(oldL7Rule, newL7Rule *akov1alpha2.L7Rule) bool {
	if oldL7Rule.ResourceVersion == newL7Rule.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldL7Rule.Spec) + oldL7Rule.Status.Status)
	newSpecHash := utils.Hash(utils.Stringify(newL7Rule.Spec) + newL7Rule.Status.Status)

	return oldSpecHash != newSpecHash
}

// SetupAKOCRDEventHandlers handles setting up of AKO CRD event handlers
// TODO: The CRD are getting re-enqueued for the same resourceVersion via fullsync as well as via these handlers.
// We can leverage the resourceVersion checks to optimize this code. However the CRDs would need a check on
// status for re-publish. The status does not change the resourceVersion and during fullsync we ignore a CRD
// if it's status is not updated.
func (c *AviController) SetupAKOCRDEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AKO CRD Event handlers")
	informer := lib.AKOControlConfig().CRDInformers()

	if lib.AKOControlConfig().HostRuleEnabled() {
		hostRuleEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				hostrule := obj.(*akov1beta1.HostRule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
				key := lib.HostRule + "/" + utils.ObjKey(hostrule)
				if err := c.GetValidator().ValidateHostRuleObj(key, hostrule); err != nil {
					utils.AviLog.Warnf("key: %s, msg: Error retrieved during validation of HostRule: %v", key, err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1beta1.HostRule)
				hostrule := new.(*akov1beta1.HostRule)
				if isHostRuleUpdated(oldObj, hostrule) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
					key := lib.HostRule + "/" + utils.ObjKey(hostrule)
					if err := c.GetValidator().ValidateHostRuleObj(key, hostrule); err != nil {
						utils.AviLog.Warnf("key: %s, Error retrieved during validation of HostRule: %v", key, err)
					}
					if oldObj.Spec.VirtualHost.L7Rule != "" && oldObj.Spec.VirtualHost.L7Rule != hostrule.Spec.VirtualHost.L7Rule {
						objects.SharedCRDLister().DeleteL7RuleToHostRuleMapping(namespace+"/"+oldObj.Spec.VirtualHost.L7Rule, oldObj.Name)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				hostrule, ok := obj.(*akov1beta1.HostRule)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					hostrule, ok = tombstone.Obj.(*akov1beta1.HostRule)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an HostRule: %#v", obj)
						return
					}
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
				key := lib.HostRule + "/" + utils.ObjKey(hostrule)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
				objects.SharedResourceVerInstanceLister().Delete(key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
		}

		informer.HostRuleInformer.Informer().AddEventHandler(hostRuleEventHandler)
	}

	if lib.AKOControlConfig().HttpRuleEnabled() {
		httpRuleEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				httprule := obj.(*akov1beta1.HTTPRule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
				key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
				if err := c.GetValidator().ValidateHTTPRuleObj(key, httprule); err != nil {
					utils.AviLog.Warnf("Error retrieved during validation of HTTPRule: %v", err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1beta1.HTTPRule)
				httprule := new.(*akov1beta1.HTTPRule)
				// reflect.DeepEqual does not work on type []byte,
				// unable to capture edits in destinationCA
				if isHTTPRuleUpdated(oldObj, httprule) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
					key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
					if err := c.GetValidator().ValidateHTTPRuleObj(key, httprule); err != nil {
						utils.AviLog.Warnf("Error retrieved during validation of HTTPRule: %v", err)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				httprule, ok := obj.(*akov1beta1.HTTPRule)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					httprule, ok = tombstone.Obj.(*akov1beta1.HTTPRule)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an HTTPRule: %#v", obj)
						return
					}
				}
				key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
				// no need to validate for delete handler
				bkt := utils.Bkt(namespace, numWorkers)
				objects.SharedResourceVerInstanceLister().Delete(key)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
		}

		informer.HTTPRuleInformer.Informer().AddEventHandler(httpRuleEventHandler)
	}

	if lib.AKOControlConfig().AviInfraSettingEnabled() {
		aviInfraEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				aviinfra := obj.(*akov1beta1.AviInfraSetting)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(aviinfra))
				key := lib.AviInfraSetting + "/" + utils.ObjKey(aviinfra)
				if err := c.GetValidator().ValidateAviInfraSetting(key, aviinfra); err != nil {
					utils.AviLog.Warnf("Error retrieved during validation of AviInfraSetting: %v", err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1beta1.AviInfraSetting)
				aviInfra := new.(*akov1beta1.AviInfraSetting)
				if isAviInfraUpdated(oldObj, aviInfra) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(aviInfra))
					key := lib.AviInfraSetting + "/" + utils.ObjKey(aviInfra)
					if err := c.GetValidator().ValidateAviInfraSetting(key, aviInfra); err != nil {
						utils.AviLog.Warnf("Error retrieved during validation of AviInfraSetting: %v", err)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				aviinfra, ok := obj.(*akov1beta1.AviInfraSetting)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					aviinfra, ok = tombstone.Obj.(*akov1beta1.AviInfraSetting)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an AviInfraSetting: %#v", obj)
						return
					}
				}
				key := lib.AviInfraSetting + "/" + utils.ObjKey(aviinfra)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(aviinfra))
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
				objects.SharedResourceVerInstanceLister().Delete(key)
				// no need to validate for delete handler
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
		}

		informer.AviInfraSettingInformer.Informer().AddEventHandler(aviInfraEventHandler)
	}

	if lib.AKOControlConfig().SsoRuleEnabled() {
		ssoRuleEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				ssoRule := obj.(*akov1alpha2.SSORule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ssoRule))
				key := lib.SSORule + "/" + utils.ObjKey(ssoRule)
				if err := c.GetValidator().ValidateSSORuleObj(key, ssoRule); err != nil {
					utils.AviLog.Warnf("key: %s, msg: Error retrieved during validation of SSORule: %v", key, err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1alpha2.SSORule)
				ssoRule := new.(*akov1alpha2.SSORule)
				if isSSORuleUpdated(oldObj, ssoRule) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ssoRule))
					key := lib.SSORule + "/" + utils.ObjKey(ssoRule)
					if err := c.GetValidator().ValidateSSORuleObj(key, ssoRule); err != nil {
						utils.AviLog.Warnf("key: %s, Error retrieved during validation of SSORule: %v", key, err)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				ssoRule, ok := obj.(*akov1alpha2.SSORule)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					ssoRule, ok = tombstone.Obj.(*akov1alpha2.SSORule)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an SSORule: %#v", obj)
						return
					}
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(ssoRule))
				key := lib.SSORule + "/" + utils.ObjKey(ssoRule)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
				objects.SharedResourceVerInstanceLister().Delete(key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
		}
		informer.SSORuleInformer.Informer().AddEventHandler(ssoRuleEventHandler)
	}

	if lib.AKOControlConfig().L4RuleEnabled() {
		l4RuleEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				l4Rule := obj.(*akov1alpha2.L4Rule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(l4Rule))
				key := lib.L4Rule + "/" + utils.ObjKey(l4Rule)
				if err := c.GetValidator().ValidateL4RuleObj(key, l4Rule); err != nil {
					utils.AviLog.Warnf("Error retrieved during validation of L4Rule: %v", err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1alpha2.L4Rule)
				l4Rule := new.(*akov1alpha2.L4Rule)
				if isL4RuleUpdated(oldObj, l4Rule) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(l4Rule))
					key := lib.L4Rule + "/" + utils.ObjKey(l4Rule)
					if err := c.GetValidator().ValidateL4RuleObj(key, l4Rule); err != nil {
						utils.AviLog.Warnf("Error retrieved during validation of L4Rule: %v", err)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
					lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				l4Rule, ok := obj.(*akov1alpha2.L4Rule)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					l4Rule, ok = tombstone.Obj.(*akov1alpha2.L4Rule)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an L4Rule: %#v", obj)
						return
					}
				}
				key := lib.L4Rule + "/" + utils.ObjKey(l4Rule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(l4Rule))
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)

				// Clean up HealthMonitor to L4Rule mappings when L4Rule is deleted
				l4RuleNsName := namespace + "/" + l4Rule.Name
				c.cleanupHealthMonitorToL4RuleMappings(key, l4RuleNsName, l4Rule)

				bkt := utils.Bkt(namespace, numWorkers)
				objects.SharedResourceVerInstanceLister().Delete(key)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			},
		}
		informer.L4RuleInformer.Informer().AddEventHandler(l4RuleEventHandler)
	}

	// HealthMonitor dynamic event handler - only when L4Rule is enabled
	if lib.AKOControlConfig().L4RuleEnabled() {
		healthMonitorEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}

				healthMonitorObj, ok := obj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Warn("Error in converting object to HealthMonitor object")
					return
				}

				namespace, name := healthMonitorObj.GetNamespace(), healthMonitorObj.GetName()
				if namespace == "" || name == "" {
					return
				}

				key := lib.HealthMonitor + "/" + namespace + "/" + name
				utils.AviLog.Debugf("key: %s, msg: ADD", key)

				// Find all L4Rules that reference this HealthMonitor and re-queue them
				// (Mapping only exists if HealthMonitor was previously processed)
				c.processL4RulesForHealthMonitor(key, namespace, name, numWorkers)
			},
			UpdateFunc: func(oldObj, curObj interface{}) {
				if c.DisableSync {
					return
				}
				oldHealthMonitorObj, ok := oldObj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Warn("Error in converting old object to HealthMonitor object")
					return
				}

				curHealthMonitorObj, ok := curObj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Warn("Error in converting current object to HealthMonitor object")
					return
				}

				// Check if resource version changed
				if oldHealthMonitorObj.GetResourceVersion() != curHealthMonitorObj.GetResourceVersion() {
					namespace, name := curHealthMonitorObj.GetNamespace(), curHealthMonitorObj.GetName()
					if namespace == "" || name == "" {
						return
					}

					key := lib.HealthMonitor + "/" + namespace + "/" + name
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)

					// Find all L4Rules that reference this HealthMonitor and re-queue them
					// (Mapping only exists if HealthMonitor was previously processed)
					c.processL4RulesForHealthMonitor(key, namespace, name, numWorkers)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}

				healthMonitorObj, ok := obj.(*unstructured.Unstructured)
				if !ok {
					// healthMonitorObj was deleted but its final state is unrecorded.
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					healthMonitorObj, ok = tombstone.Obj.(*unstructured.Unstructured)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not a HealthMonitor: %#v", obj)
						return
					}
				}

				namespace, name := healthMonitorObj.GetNamespace(), healthMonitorObj.GetName()
				if namespace == "" || name == "" {
					return
				}
				key := lib.HealthMonitor + "/" + namespace + "/" + name
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)

				// Find all L4Rules that reference this HealthMonitor and re-queue them
				// (Mapping only exists if HealthMonitor was previously processed)
				c.processL4RulesForHealthMonitor(key, namespace, name, numWorkers)

				// Note: We do NOT delete the HealthMonitor-to-L4Rule mapping here.
				// The mapping should persist even when HealthMonitor is deleted because:
				// 1. L4Rule still exists and still references the deleted HealthMonitor
				// 2. If HealthMonitor is recreated, L4Rule should be re-evaluated
				// 3. Mapping cleanup happens only when L4Rule itself is deleted
			},
		}

		// Add event handler to HealthMonitor dynamic informer only if AKO CRD Operator is enabled
		if !lib.IsAKOCRDOperatorEnabled() {
			utils.AviLog.Warnf("Skipping HealthMonitor event handler setup as AKO CRD Operator is not enabled")
		} else if c.dynamicInformers != nil && c.dynamicInformers.HealthMonitorInformer != nil {
			c.dynamicInformers.HealthMonitorInformer.Informer().AddEventHandler(healthMonitorEventHandler)
		}
	}

	if lib.AKOControlConfig().L7RuleEnabled() {
		l7RuleEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				l7Rule := obj.(*akov1alpha2.L7Rule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(l7Rule))
				key := lib.L7Rule + "/" + utils.ObjKey(l7Rule)
				if err := c.GetValidator().ValidateL7RuleObj(key, l7Rule); err != nil {
					utils.AviLog.Warnf("Error retrieved during validation of L7Rule: %v", err)
					return
				}
				utils.AviLog.Debugf("key: %s, msg: Add", key)
				found, hostRules := objects.SharedCRDLister().GetL7RuleToHostRuleMapping(namespace + "/" + l7Rule.Name)
				if found {
					for hr := range hostRules {
						hostrule, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hr)
						if err != nil {
							utils.AviLog.Warnf("key: %s, msg: HostRule %s not found for L7Rule msg: %v", key, hr, err)
							continue
						}
						key := lib.HostRule + "/" + utils.ObjKey(hostrule)
						utils.AviLog.Debugf("key: %s, msg: Update", key)
						bkt := utils.Bkt(namespace, numWorkers)
						c.workqueue[bkt].AddRateLimited(key)
						lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
					}
				}
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1alpha2.L7Rule)
				l7Rule := new.(*akov1alpha2.L7Rule)
				if isL7RuleUpdated(oldObj, l7Rule) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(l7Rule))
					key := lib.L7Rule + "/" + utils.ObjKey(l7Rule)
					if err := c.GetValidator().ValidateL7RuleObj(key, l7Rule); err != nil {
						utils.AviLog.Warnf("Error retrieved during validation of L7Rule: %v", err)
						return
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					found, hostRules := objects.SharedCRDLister().GetL7RuleToHostRuleMapping(namespace + "/" + l7Rule.Name)
					if found {
						for hr := range hostRules {
							hostrule, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hr)
							if err != nil {
								utils.AviLog.Warnf("key: %s, msg: HostRule %s not found for L7Rule msg: %v", key, hr, err)
								continue
							}
							key := lib.HostRule + "/" + utils.ObjKey(hostrule)
							utils.AviLog.Debugf("key: %s, msg: Update", key)
							bkt := utils.Bkt(namespace, numWorkers)
							c.workqueue[bkt].AddRateLimited(key)
							lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
						}
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				l7Rule, ok := obj.(*akov1alpha2.L7Rule)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					l7Rule, ok = tombstone.Obj.(*akov1alpha2.L7Rule)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an L7Rule: %#v", obj)
						return
					}
				}
				key := lib.L7Rule + "/" + utils.ObjKey(l7Rule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(l7Rule))
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
				objects.SharedResourceVerInstanceLister().Delete(key)
				found, hostRules := objects.SharedCRDLister().GetL7RuleToHostRuleMapping(namespace + "/" + l7Rule.Name)
				if found {
					for hr := range hostRules {
						hostrule, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(namespace).Get(hr)
						if err != nil {
							utils.AviLog.Warnf("key: %s, msg: HostRule %s not found for L7Rule msg: %v", key, hr, err)
							continue
						}
						key := lib.HostRule + "/" + utils.ObjKey(hostrule)
						utils.AviLog.Debugf("key: %s, msg: Update", key)
						bkt := utils.Bkt(namespace, numWorkers)
						c.workqueue[bkt].AddRateLimited(key)
						lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
					}
				}
			},
		}
		informer.L7RuleInformer.Informer().AddEventHandler(l7RuleEventHandler)
	}
}

// SetupIstioCRDEventHandlers handles setting up of Istio CRD event handlers
func (c *AviController) SetupIstioCRDEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AKO Istio CRD Event handlers")
	informer := lib.AKOControlConfig().IstioCRDInformers()

	virtualServiceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			vs := obj.(*istiov1alpha3.VirtualService)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
			key := lib.IstioVirtualService + "/" + utils.ObjKey(vs)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == vs.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: Same resource version returning", key)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*istiov1alpha3.VirtualService)
			vs := new.(*istiov1alpha3.VirtualService)
			if !reflect.DeepEqual(oldObj.Spec, vs.Spec) { //nolint:govet
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
				key := lib.IstioVirtualService + "/" + utils.ObjKey(vs)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			vs, ok := obj.(*istiov1alpha3.VirtualService)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				vs, ok = tombstone.Obj.(*istiov1alpha3.VirtualService)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an vs: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
			key := lib.IstioVirtualService + "/" + utils.ObjKey(vs)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			objects.SharedResourceVerInstanceLister().Delete(key)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
	}

	informer.VirtualServiceInformer.Informer().AddEventHandler(virtualServiceEventHandler)

	drEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			dr := obj.(*istiov1alpha3.DestinationRule)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(dr))
			key := lib.IstioDestinationRule + "/" + utils.ObjKey(dr)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == dr.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: Same resource version returning", key)
				return
			}
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*istiov1alpha3.DestinationRule)
			dr := new.(*istiov1alpha3.DestinationRule)
			if !reflect.DeepEqual(oldObj.Spec, dr.Spec) { //nolint:govet
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(dr))
				key := lib.IstioDestinationRule + "/" + utils.ObjKey(dr)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			dr, ok := obj.(*istiov1alpha3.DestinationRule)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				dr, ok = tombstone.Obj.(*istiov1alpha3.DestinationRule)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an vs: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(dr))
			key := lib.IstioDestinationRule + "/" + utils.ObjKey(dr)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			objects.SharedResourceVerInstanceLister().Delete(key)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
	}

	informer.DestinationRuleInformer.Informer().AddEventHandler(drEventHandler)

	gatewayEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			vs := obj.(*istiov1alpha3.Gateway)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
			key := lib.IstioGateway + "/" + utils.ObjKey(vs)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			ok, resVer := objects.SharedResourceVerInstanceLister().Get(key)
			if ok && resVer.(string) == vs.ResourceVersion {
				utils.AviLog.Debugf("key: %s, msg: Same resource version returning", key)
				return
			}
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*istiov1alpha3.Gateway)
			vs := new.(*istiov1alpha3.Gateway)
			if !reflect.DeepEqual(oldObj.Spec, vs.Spec) { //nolint:govet
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
				key := lib.IstioGateway + "/" + utils.ObjKey(vs)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			vs, ok := obj.(*istiov1alpha3.Gateway)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				vs, ok = tombstone.Obj.(*istiov1alpha3.Gateway)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an vs: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
			key := lib.IstioGateway + "/" + utils.ObjKey(vs)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			objects.SharedResourceVerInstanceLister().Delete(key)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
	}

	informer.GatewayInformer.Informer().AddEventHandler(gatewayEventHandler)

}

// SetupMultiClusterIngressEventHandlers handles setting up of MultiClusterIngress CRD event handlers
func (c *AviController) SetupMultiClusterIngressEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up MultiClusterIngress CRD Event handlers")

	multiClusterIngressEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			mci := obj.(*akov1alpha1.MultiClusterIngress)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(mci))
			key := lib.MultiClusterIngress + "/" + utils.ObjKey(mci)
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Multi-cluster Ingress add event: Namespace: %s didn't qualify filter. Not adding multi-cluster ingress", key, namespace)
				return
			}
			if err := c.GetValidator().ValidateMultiClusterIngressObj(key, mci); err != nil {
				utils.AviLog.Warnf("key: %s, msg: Validation of MultiClusterIngress failed: %v", key, err)
				return
			}
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*akov1alpha1.MultiClusterIngress)
			mci := new.(*akov1alpha1.MultiClusterIngress)
			if !reflect.DeepEqual(oldObj.Spec, mci.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(mci))
				key := lib.MultiClusterIngress + "/" + utils.ObjKey(mci)
				if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Multi-cluster Ingress update event: Namespace: %s didn't qualify filter. Not updating multi-cluster ingress", key, namespace)
					return
				}
				if err := c.GetValidator().ValidateMultiClusterIngressObj(key, mci); err != nil {
					utils.AviLog.Warnf("key: %s, msg: Validation of MultiClusterIngress failed: %v", key, err)
					return
				}
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			mci, ok := obj.(*akov1alpha1.MultiClusterIngress)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				mci, ok = tombstone.Obj.(*akov1alpha1.MultiClusterIngress)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a MultiClusterIngress: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(mci))
			key := lib.MultiClusterIngress + "/" + utils.ObjKey(mci)
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Multi-cluster Ingress delete event: Namespace: %s didn't qualify filter. Not deleting multi-cluster ingress", key, namespace)
				return
			}
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			objects.SharedResourceVerInstanceLister().Delete(key)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
	}
	c.informers.MultiClusterIngressInformer.Informer().AddEventHandler(multiClusterIngressEventHandler)
}

// SetupServiceImportEventHandlers handles setting up of ServiceImport CRD event handlers
func (c *AviController) SetupServiceImportEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up ServiceImport CRD Event handlers")

	serviceImportEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			si := obj.(*akov1alpha1.ServiceImport)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(si))
			key := lib.ServiceImport + "/" + utils.ObjKey(si)
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Service Import add event: Namespace: %s didn't qualify filter. Not adding Service Import", key, namespace)
				return
			}
			if err := c.GetValidator().ValidateServiceImportObj(key, si); err != nil {
				utils.AviLog.Warnf("key: %s, msg: Validation of ServiceImport failed: %v", key, err)
				return
			}
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*akov1alpha1.ServiceImport)
			si := new.(*akov1alpha1.ServiceImport)
			if !reflect.DeepEqual(oldObj.Spec, si.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(si))
				key := lib.ServiceImport + "/" + utils.ObjKey(si)
				if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
					utils.AviLog.Debugf("key: %s, msg: Service Import update event: Namespace: %s didn't qualify filter. Not updating Service Import", key, namespace)
					return
				}
				if err := c.GetValidator().ValidateServiceImportObj(key, si); err != nil {
					utils.AviLog.Warnf("key: %s, msg: Validation of ServiceImport failed: %v", key, err)
					return
				}
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
				lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			si, ok := obj.(*akov1alpha1.ServiceImport)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				si, ok = tombstone.Obj.(*akov1alpha1.ServiceImport)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a ServiceImport: %#v", obj)
					return
				}
			}
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(si))
			key := lib.ServiceImport + "/" + utils.ObjKey(si)
			if lib.IsNamespaceBlocked(namespace) || !utils.CheckIfNamespaceAccepted(namespace) {
				utils.AviLog.Debugf("key: %s, msg: Service Import delete event: Namespace: %s didn't qualify filter. Not deleting Service Import", key, namespace)
				return
			}
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			objects.SharedResourceVerInstanceLister().Delete(key)
			c.workqueue[bkt].AddRateLimited(key)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		},
	}
	c.informers.ServiceImportInformer.Informer().AddEventHandler(serviceImportEventHandler)
}

func checkRefsOnController(key string, refMap map[string]string, tenant string) error {
	for k, value := range refMap {
		if k == "" {
			continue
		}

		if err := checkRefOnController(key, value, k, tenant); err != nil {
			return err
		}
	}
	return nil
}

var refModelMap = map[string]string{
	"SslKeyCert":             "sslkeyandcertificate",
	"WafPolicy":              "wafpolicy",
	"HttpPolicySet":          "httppolicyset",
	"SslProfile":             "sslprofile",
	"AppProfile":             "applicationprofile",
	"AnalyticsProfile":       "analyticsprofile",
	"ErrorPageProfile":       "errorpageprofile",
	"VsDatascript":           "vsdatascriptset",
	"HealthMonitor":          "healthmonitor",
	"ApplicationPersistence": "applicationpersistenceprofile",
	"PKIProfile":             "pkiprofile",
	"ServiceEngineGroup":     "serviceenginegroup",
	"Network":                "network",
	"NetworkUUID":            "network",
	"SSOPolicy":              "ssopolicy",
	"AuthProfile":            "authprofile",
	"ICAPProfile":            "icapprofile",
	"NetworkProfile":         "networkprofile",
	"SecurityPolicy":         "securitypolicy",
	"NetworkSecurityPolicy":  "networksecuritypolicy",
	"BotPolicy":              "botdetectionpolicy",
	"TrafficCloneProfile":    "trafficcloneprofile",
}

// checkRefOnController checks whether a provided ref on the controller
func checkRefOnController(key, refKey, refValue, tenant string) error {
	// assign the last avi client for ref checks
	aviClientLen := lib.GetshardSize()
	clients := avicache.SharedAVIClients(tenant)
	uri := fmt.Sprintf("/api/%s?name=%s&fields=name,type,labels,created_by", refModelMap[refKey], refValue)

	// For public clouds, check using network UUID in AWS, normal network API for GCP, skip altogether for Azure.
	// If reference key is network uuid , then check using UUID.
	if (lib.IsPublicCloud() && refModelMap[refKey] == "network") || refKey == "NetworkUUID" {
		if lib.UsesNetworkRef() || refKey == "NetworkUUID" {
			// During the portal-webapp migration from Python to Go, network views were not correctly ported. However, network APIs are now being routed through Go code,
			// which is incorrect. The Avi Controller needs to be fixed.
			// For now, subnet UUID validation is disabled for AWS and Azure clouds to avoid impact on EKS and AKS deployments.
			cloudType := lib.GetCloudType()
			if cloudType == lib.CLOUD_AWS || cloudType == lib.CLOUD_AZURE {
				utils.AviLog.Infof("Cloud Type is %q, skip validating references on controller", cloudType)
				return nil
			}
			var rest_response interface{}
			utils.AviLog.Infof("Cloud is  %s, checking network ref using uuid", lib.GetCloudType())
			uri := fmt.Sprintf("/api/%s/%s?cloud_uuid=%s", refModelMap[refKey], refValue, lib.GetCloudUUID())
			err := lib.AviGet(clients.AviClient[aviClientLen], uri, &rest_response)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: Get uri %v returned err %v", key, uri, err)
				return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
			} else if rest_response != nil {
				utils.AviLog.Infof("Found %s %s on controller", refModelMap[refKey], refValue)
				return nil
			} else {
				utils.AviLog.Warnf("key: %s, msg: No Objects found for refName: %s/%s", key, refModelMap[refKey], refValue)
				return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
			}
		}
	}

	result, err := lib.AviGetCollectionRaw(clients.AviClient[aviClientLen], uri)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Get uri %v returned err %v", key, uri, err)
		return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	if result.Count == 0 {
		utils.AviLog.Warnf("key: %s, msg: No Objects found for refName: %s/%s", key, refModelMap[refKey], refValue)
		return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	items := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &items)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal results, err: %v", key, err)
		return fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	item := make(map[string]interface{})
	err = json.Unmarshal(items[0], &item)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal item, err: %v", key, err)
		return fmt.Errorf("%s \"%s\" found on controller is invalid", refModelMap[refKey], refValue)
	}

	switch refKey {
	case "AppProfile":
		if appProfType, ok := item["type"].(string); ok {
			objType, _, _ := lib.ExtractTypeNameNamespace(key)
			if objType == lib.L4Rule {
				if appProfType != lib.AllowedL4ApplicationProfile {
					utils.AviLog.Warnf("key: %s, msg: L4 applicationProfile: %s must be of type %s", key, refValue, lib.AllowedL4ApplicationProfile)
					return fmt.Errorf("%s \"%s\" found on controller is invalid, must be of type: %s",
						refModelMap[refKey], refValue, lib.AllowedL4ApplicationProfile)
				}
				return nil
			}
			if appProfType != lib.AllowedL7ApplicationProfile {
				utils.AviLog.Warnf("key: %s, msg: applicationProfile: %s must be of type %s", key, refValue, lib.AllowedL7ApplicationProfile)
				return fmt.Errorf("%s \"%s\" found on controller is invalid, must be of type: %s",
					refModelMap[refKey], refValue, lib.AllowedL7ApplicationProfile)
			}
		}
	case "ServiceEngineGroup":
		if seGroupLabels, ok := item["labels"].([]map[string]string); ok {
			if len(seGroupLabels) == 0 {
				utils.AviLog.Infof("key: %s, msg: ServiceEngineGroup %s not configured with labels", key, item["name"].(string))
			} else {
				if !reflect.DeepEqual(seGroupLabels, lib.GetLabels()) {
					utils.AviLog.Warnf("key: %s, msg: serviceEngineGroup: %s mismatched labels %s", key, refValue, utils.Stringify(seGroupLabels))
					return fmt.Errorf("%s \"%s\" found on controller is invalid, mismatched labels: %s",
						refModelMap[refKey], refValue, utils.Stringify(seGroupLabels))
				}
			}
		}
	}

	if itemCreatedBy, ok := item["created_by"].(string); ok && itemCreatedBy == lib.GetAKOUser() {
		utils.AviLog.Warnf("key: %s, msg: Cannot use object referred in CRD created by current AKO instance", key)
		return fmt.Errorf("%s \"%s\" Invalid operation, object referred is created by current AKO instance",
			refModelMap[refKey], refValue)
	}

	utils.AviLog.Infof("key: %s, msg: Ref found for %s/%s", key, refModelMap[refKey], refValue)
	return nil
}

// checkForL4SSLAppProfile checks if the app profile specified in l4rule is of type L4 or L4 SSL.
// If app profile is of type L4 SSL it returns true and nil.
// If app profile is of type L4 it returns false and nil.
// Otherwise false and specific error is returned.
func checkForL4SSLAppProfile(key, refValue string) (bool, error) {
	// assign the last avi client for ref checks
	refKey := "AppProfile"
	aviClientLen := lib.GetshardSize()
	clients := avicache.SharedAVIClients(lib.GetTenant())
	uri := fmt.Sprintf("/api/%s?name=%s&fields=name,type,labels,created_by", refModelMap[refKey], refValue)

	result, err := lib.AviGetCollectionRaw(clients.AviClient[aviClientLen], uri)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Get uri %v returned err %v", key, uri, err)
		return false, fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	if result.Count == 0 {
		utils.AviLog.Warnf("key: %s, msg: No Objects found for refName: %s/%s", key, refModelMap[refKey], refValue)
		return false, fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	items := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &items)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal results, err: %v", key, err)
		return false, fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}
	item := make(map[string]interface{})
	err = json.Unmarshal(items[0], &item)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal item, err: %v", key, err)
		return false, fmt.Errorf("%s \"%s\" found on controller is invalid", refModelMap[refKey], refValue)
	}
	if appProfType, ok := item["type"].(string); ok {
		if appProfType == lib.AllowedL4SSLApplicationProfile {
			utils.AviLog.Infof("key: %s, msg: Ref found for %s/%s", key, refModelMap[refKey], refValue)
			return true, nil
		}
		if appProfType == lib.AllowedL4ApplicationProfile {
			utils.AviLog.Infof("key: %s, msg: Ref found for %s/%s", key, refModelMap[refKey], refValue)
			return false, nil
		}
	}
	utils.AviLog.Warnf("key: %s, msg: L4 applicationProfile: %s must be of type %s or %s", key, refValue, lib.AllowedL4ApplicationProfile, lib.AllowedL4SSLApplicationProfile)
	return false, fmt.Errorf("%s \"%s\" found on controller is invalid, must be of type: %s or %s",
		refModelMap[refKey], refValue, lib.AllowedL4ApplicationProfile, lib.AllowedL4SSLApplicationProfile)
}

// checkForNetworkProfileTypeTCP checks if the network profile specified in l4rule is of type TCP proxy.
// If network profile is of type TCP proxy it returns true and nil.
// Otherwise false and specific error is returned.
func checkForNetworkProfileTypeTCP(key, refValue string) (bool, error) {
	// assign the last avi client for ref checks
	refKey := "NetworkProfile"
	aviClientLen := lib.GetshardSize()
	clients := avicache.SharedAVIClients(lib.GetTenant())
	uri := fmt.Sprintf("/api/%s?name=%s&fields=name,profile,labels,created_by", refModelMap[refKey], refValue)

	result, err := lib.AviGetCollectionRaw(clients.AviClient[aviClientLen], uri)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Get uri %v returned err %v", key, uri, err)
		return false, fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	if result.Count == 0 {
		utils.AviLog.Warnf("key: %s, msg: No Objects found for refName: %s/%s", key, refModelMap[refKey], refValue)
		return false, fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	items := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &items)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal results, err: %v", key, err)
		return false, fmt.Errorf("%s \"%s\" not found on controller", refModelMap[refKey], refValue)
	}

	item := make(map[string]interface{})
	err = json.Unmarshal(items[0], &item)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Failed to unmarshal item, err: %v", key, err)
		return false, fmt.Errorf("%s \"%s\" found on controller is invalid", refModelMap[refKey], refValue)
	}
	if profile, ok := item["profile"].(map[string]interface{}); ok {
		if networkProfType, ok := profile["type"].(string); ok && networkProfType == lib.AllowedTCPProxyNetworkProfileType {
			return true, nil
		}
	}
	utils.AviLog.Warnf("key: %s, msg: Network profile : %s must be of type %s for L4 SSL support", key, refValue, lib.AllowedTCPProxyNetworkProfileType)
	return false, fmt.Errorf("%s \"%s\" found on controller is invalid, must be of type: %s for L4 SSL support",
		refModelMap[refKey], refValue, lib.AllowedTCPProxyNetworkProfileType)
}

// addSeGroupLabel configures SEGroup with appropriate labels, during AviInfraSetting
// creation/updates after ingestion
func addSeGroupLabel(key, segName string) {
	// No need to configure labels if static route sync is disabled globally.
	if lib.GetDisableStaticRoute() {
		utils.AviLog.Infof("Skipping the check for SE group labels for SEG %s", segName)
		return
	}

	// assign the last avi client for ref checks
	clients := avicache.SharedAVIClients(lib.GetTenant())
	aviClientLen := lib.GetshardSize()

	// configure labels on SeGroup if not present already.
	seGroup, err := avicache.GetAviSeGroup(clients.AviClient[aviClientLen], segName)
	if err != nil {
		utils.AviLog.Errorf("Failed to get SE group")
		return
	}

	avicache.ConfigureSeGroupLabels(clients.AviClient[aviClientLen], seGroup)
}

func SetAviInfrasettingVIPNetworks(name, segMgmtNetwork, infraSEGName string, netAviInfra []akov1beta1.AviInfraSettingVipNetwork) {
	// assign the last avi client for ref checks
	clients := avicache.SharedAVIClients(lib.GetTenant())
	aviClientLen := lib.GetshardSize()
	network := netAviInfra
	var err error
	if lib.GetCloudType() == lib.CLOUD_VCENTER || lib.GetCloudType() == lib.CLOUD_NONE {
		// SEG mgmt network is required to find out host overlap. Not applicable for No Access cloud.
		if lib.GetCloudType() == lib.CLOUD_VCENTER && infraSEGName == "" && segMgmtNetwork == "" {
			segMgmtNetwork = avicache.GetCMSEGManagementNetwork(clients.AviClient[aviClientLen])
		}
		network, err = avicache.PopulateVipNetworkwithUUID(segMgmtNetwork, clients.AviClient[aviClientLen], netAviInfra)
		if len(network) == 0 {
			utils.AviLog.Errorf("Infrasetting: %s not applied, Error occurred while populating vip network list. Err: %s", name, err.Error())
			// Need to check this return
			return
		}
	}
	utils.AviLog.Debugf("Infrasetting: %s, VIP Network Obtained in AviInfrasetting: %v", name, utils.Stringify(network))
	//set infrasetting name specific vip network
	lib.SetVipInfraNetworkList(name, network)
}

func SetAviInfrasettingNodeNetworks(name, segMgmtNetwork, infraSEGName string, netAviInfra []akov1beta1.AviInfraSettingNodeNetwork) {
	// assign the last avi client for ref checks
	clients := avicache.SharedAVIClients(lib.GetTenant())
	aviClientLen := lib.GetshardSize()
	nodeNetorkList := make(map[string]lib.NodeNetworkMap)
	var err error

	for _, net := range netAviInfra {
		nwMap := lib.NodeNetworkMap{
			Cidrs: net.Cidrs,
		}
		// Give preference to networkUUID
		if net.NetworkUUID != "" {
			nwMap.NetworkUUID = net.NetworkUUID
			nodeNetorkList[net.NetworkUUID] = nwMap
		} else if net.NetworkName != "" {
			nodeNetorkList[net.NetworkName] = nwMap
		}
	}

	if lib.GetCloudType() == lib.CLOUD_VCENTER || lib.GetCloudType() == lib.CLOUD_NONE {
		if lib.GetCloudType() == lib.CLOUD_VCENTER && infraSEGName == "" && segMgmtNetwork == "" {
			segMgmtNetwork = avicache.GetCMSEGManagementNetwork(clients.AviClient[aviClientLen])
		}
		ret := avicache.FetchNodeNetworks(segMgmtNetwork, clients.AviClient[aviClientLen], &err, nodeNetorkList)
		if !ret {
			utils.AviLog.Infof("Infrasetting: %s is not applied, Error occurred: %s", name, err.Error())
			return
		}
	}
	utils.AviLog.Debugf("Infrasetting: %s Node Network Obtained in AviInfrasetting: %v", name, utils.Stringify(nodeNetorkList))
	//set infrasetting name specific node network
	lib.SetNodeInfraNetworkList(name, nodeNetorkList)
}

// Fetch SEG mgmt network
func GetSEGManagementNetwork(name string) string {
	mgmtNetwork := ""
	// assign the last avi client for ref checks
	clients := avicache.SharedAVIClients(lib.GetTenant())
	aviClientLen := lib.GetshardSize()
	seg, err := avicache.GetAviSeGroup(clients.AviClient[aviClientLen], name)
	if err == nil {
		// seg MgmtNetwork ref contains network-uuid based url.
		if seg.MgmtNetworkRef != nil {
			parts := strings.Split(*seg.MgmtNetworkRef, "/")
			mgmtNetwork = parts[len(parts)-1]
		}
	}
	return mgmtNetwork
}

func (c *AviController) SyncCRDObjects() {
	utils.AviLog.Debugf("Starting syncing all CRD objects")

	l7RuleObjs, err := lib.AKOControlConfig().CRDInformers().L7RuleInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the L7Rules during full sync: %s", err)
	} else {
		for _, l7Rule := range l7RuleObjs {
			key := lib.L7Rule + "/" + utils.ObjKey(l7Rule)
			if err := c.GetValidator().ValidateL7RuleObj(key, l7Rule); err != nil {
				utils.AviLog.Warnf("key: %s, Error during validation of L7Rule: %v", key, err)
			}
		}
	}

	hostRuleObjs, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the hostrules during full sync: %s", err)
	} else {
		for _, hostRuleObj := range hostRuleObjs {
			key := lib.HostRule + "/" + utils.ObjKey(hostRuleObj)
			if err := c.GetValidator().ValidateHostRuleObj(key, hostRuleObj); err != nil {
				utils.AviLog.Warnf("key: %s, Error during validation of HostRule: %v", key, err)
			}
		}
	}

	httpRuleObjs, err := lib.AKOControlConfig().CRDInformers().HTTPRuleInformer.Lister().HTTPRules(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the httprules during full sync: %s", err)
	} else {
		for _, httpRuleObj := range httpRuleObjs {
			key := lib.HTTPRule + "/" + utils.ObjKey(httpRuleObj)
			if err := c.GetValidator().ValidateHTTPRuleObj(key, httpRuleObj); err != nil {
				utils.AviLog.Warnf("key: %s, Error during validation of HTTPRule: %v", key, err)
			}
		}
	}

	aviInfraObjs, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the avinfrasettings during full sync: %s", err)
	} else {
		for _, aviInfraObj := range aviInfraObjs {
			key := lib.AviInfraSetting + "/" + utils.ObjKey(aviInfraObj)
			if err := c.GetValidator().ValidateAviInfraSetting(key, aviInfraObj); err != nil {
				utils.AviLog.Warnf("key: %s, Error during validation of AviInfraSetting: %v", key, err)
			}
		}
	}

	ssoRuleObjs, err := lib.AKOControlConfig().CRDInformers().SSORuleInformer.Lister().SSORules(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the SsoRules during full sync: %s", err)
	} else {
		for _, ssoRuleObj := range ssoRuleObjs {
			key := lib.SSORule + "/" + utils.ObjKey(ssoRuleObj)
			if err := c.GetValidator().ValidateSSORuleObj(key, ssoRuleObj); err != nil {
				utils.AviLog.Warnf("key: %s, Error during validation of SSORule : %v", key, err)
			}
		}
	}

	l4RuleObjs, err := lib.AKOControlConfig().CRDInformers().L4RuleInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the L4Rules during full sync: %s", err)
	} else {
		for _, l4Rule := range l4RuleObjs {
			key := lib.L4Rule + "/" + utils.ObjKey(l4Rule)
			if err := c.GetValidator().ValidateL4RuleObj(key, l4Rule); err != nil {
				utils.AviLog.Warnf("key: %s, Error during validation of L4Rule: %v", key, err)
			}
		}
	}
	utils.AviLog.Debugf("Successfully synced all CRD objects")
}

// cleanupHealthMonitorToL4RuleMappings removes all HealthMonitor to L4Rule mappings for a deleted L4Rule
func (c *AviController) cleanupHealthMonitorToL4RuleMappings(key, l4RuleNsName string, l4Rule *akov1alpha2.L4Rule) {
	utils.AviLog.Debugf("key: %s, msg: Cleaning up HealthMonitor mappings for deleted L4Rule %s", key, l4RuleNsName)

	// Extract HealthMonitor references from the deleted L4Rule's healthMonitorCrdRefs
	if l4Rule.Spec.BackendProperties != nil {
		for _, backendProperty := range l4Rule.Spec.BackendProperties {
			for _, healthMonitorName := range backendProperty.HealthMonitorCrdRefs {
				if healthMonitorName != "" {
					// Remove this L4Rule from the HealthMonitor mapping
					healthMonitorNsName := l4Rule.Namespace + "/" + healthMonitorName
					objects.SharedCRDLister().DeleteHealthMonitorToL4RuleMapping(healthMonitorNsName, l4RuleNsName)
					utils.AviLog.Infof("key: %s, msg: Removed L4Rule %s from HealthMonitor %s mapping", key, l4RuleNsName, healthMonitorNsName)
				}
			}
		}
	}

	utils.AviLog.Debugf("key: %s, msg: HealthMonitor mapping cleanup completed for L4Rule %s", key, l4RuleNsName)
}

// processL4RulesForHealthMonitor finds all L4Rules that reference the given HealthMonitor and re-queues them
func (c *AviController) processL4RulesForHealthMonitor(key, namespace, healthMonitorName string, numWorkers uint32) {
	healthMonitorNsName := namespace + "/" + healthMonitorName
	found, l4Rules := objects.SharedCRDLister().GetHealthMonitorToL4RuleMapping(healthMonitorNsName)
	if found {
		for l4RuleNsName := range l4Rules {
			// Parse namespace and name from l4RuleNsName
			parts := strings.Split(l4RuleNsName, "/")
			if len(parts) != 2 {
				utils.AviLog.Warnf("key: %s, msg: Invalid L4Rule namespace/name format: %s", key, l4RuleNsName)
				continue
			}
			l4RuleNamespace, l4RuleName := parts[0], parts[1]

			// Check if L4Rule still exists and get the object for validation
			l4RuleObj, err := lib.AKOControlConfig().CRDInformers().L4RuleInformer.Lister().L4Rules(l4RuleNamespace).Get(l4RuleName)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: L4Rule %s not found, removing from mapping", key, l4RuleNsName)
				objects.SharedCRDLister().DeleteHealthMonitorToL4RuleMapping(healthMonitorNsName, l4RuleNsName)
				continue
			}

			// Validate L4Rule before queuing (important for previously rejected L4Rules)
			l4RuleKey := lib.L4Rule + "/" + l4RuleNsName
			if err := c.GetValidator().ValidateL4RuleObj(l4RuleKey, l4RuleObj); err != nil {
				utils.AviLog.Warnf("key: %s, msg: L4Rule %s validation failed, not queuing: %v", key, l4RuleKey, err)
				continue
			}

			// Re-queue the L4Rule for processing only if validation passes
			utils.AviLog.Debugf("key: %s, msg: Re-queuing L4Rule %s due to HealthMonitor change (validation passed)", key, l4RuleKey)
			bkt := utils.Bkt(l4RuleNamespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(l4RuleKey)
			lib.IncrementQueueCounter(utils.ObjectIngestionLayer)
		}
	}
}

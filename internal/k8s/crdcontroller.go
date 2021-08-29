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
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"time"

	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned"
	akoinformers "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/informers/externalversions"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"k8s.io/client-go/tools/cache"
)

func NewCRDInformers(cs akocrd.Interface) {
	var akoInformerFactory akoinformers.SharedInformerFactory

	akoInformerFactory = akoinformers.NewSharedInformerFactoryWithOptions(cs, time.Second*30)
	hostRuleInformer := akoInformerFactory.Ako().V1alpha1().HostRules()
	httpRuleInformer := akoInformerFactory.Ako().V1alpha1().HTTPRules()
	aviSettingsInformer := akoInformerFactory.Ako().V1alpha1().AviInfraSettings()

	lib.SetCRDInformers(&lib.AKOCrdInformers{
		HostRuleInformer:        hostRuleInformer,
		HTTPRuleInformer:        httpRuleInformer,
		AviInfraSettingInformer: aviSettingsInformer,
	})
}

func isHTTPRuleUpdated(oldHTTPRule, newHTTPRule *akov1alpha1.HTTPRule) bool {
	if oldHTTPRule.ResourceVersion == newHTTPRule.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldHTTPRule.Spec))
	newSpecHash := utils.Hash(utils.Stringify(newHTTPRule.Spec))

	if oldSpecHash != newSpecHash {
		return true
	}

	return false
}

func isAviInfraUpdated(oldAviInfra, newAviInfra *akov1alpha1.AviInfraSetting) bool {
	if oldAviInfra.ResourceVersion == newAviInfra.ResourceVersion {
		return false
	}

	oldSpecHash := utils.Hash(utils.Stringify(oldAviInfra.Spec))
	newSpecHash := utils.Hash(utils.Stringify(newAviInfra.Spec))

	if oldSpecHash != newSpecHash {
		return true
	}

	return false
}

// SetupAKOCRDEventHandlers handles setting up of AKO CRD event handlers
func (c *AviController) SetupAKOCRDEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AKO CRD Event handlers")
	informer := lib.GetCRDInformers()

	if lib.GetHostRuleEnabled() {
		hostRuleEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				hostrule := obj.(*akov1alpha1.HostRule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
				key := lib.HostRule + "/" + utils.ObjKey(hostrule)
				if err := validateHostRuleObj(key, hostrule); err != nil {
					utils.AviLog.Warnf("Error retrieved during validation of HostRule: %v", err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1alpha1.HostRule)
				hostrule := new.(*akov1alpha1.HostRule)
				if !reflect.DeepEqual(oldObj.Spec, hostrule.Spec) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
					key := lib.HostRule + "/" + utils.ObjKey(hostrule)
					if err := validateHostRuleObj(key, hostrule); err != nil {
						utils.AviLog.Warnf("Error retrieved during validation of HostRule: %v", err)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				hostrule, ok := obj.(*akov1alpha1.HostRule)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					hostrule, ok = tombstone.Obj.(*akov1alpha1.HostRule)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an HostRule: %#v", obj)
						return
					}
				}
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
				key := lib.HostRule + "/" + utils.ObjKey(hostrule)
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
		}

		informer.HostRuleInformer.Informer().AddEventHandler(hostRuleEventHandler)
	}

	if lib.GetHttpRuleEnabled() {
		httpRuleEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				httprule := obj.(*akov1alpha1.HTTPRule)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
				key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
				if err := validateHTTPRuleObj(key, httprule); err != nil {
					utils.AviLog.Warnf("Error retrieved during validation of HTTPRule: %v", err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1alpha1.HTTPRule)
				httprule := new.(*akov1alpha1.HTTPRule)
				// reflect.DeepEqual does not work on type []byte,
				// unable to capture edits in destinationCA
				if isHTTPRuleUpdated(oldObj, httprule) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
					key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
					if err := validateHTTPRuleObj(key, httprule); err != nil {
						utils.AviLog.Warnf("Error retrieved during validation of HTTPRule: %v", err)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				httprule, ok := obj.(*akov1alpha1.HTTPRule)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					httprule, ok = tombstone.Obj.(*akov1alpha1.HTTPRule)
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
				c.workqueue[bkt].AddRateLimited(key)
			},
		}

		informer.HTTPRuleInformer.Informer().AddEventHandler(httpRuleEventHandler)
	}

	if lib.GetAviInfraSettingEnabled() {
		aviInfraEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				aviinfra := obj.(*akov1alpha1.AviInfraSetting)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(aviinfra))
				key := lib.AviInfraSetting + "/" + utils.ObjKey(aviinfra)
				if err := validateAviInfraSetting(key, aviinfra); err != nil {
					utils.AviLog.Warnf("Error retrieved during validation of AviInfraSetting: %v", err)
				}
				utils.AviLog.Debugf("key: %s, msg: ADD", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
			UpdateFunc: func(old, new interface{}) {
				if c.DisableSync {
					return
				}
				oldObj := old.(*akov1alpha1.AviInfraSetting)
				aviInfra := new.(*akov1alpha1.AviInfraSetting)
				if isAviInfraUpdated(oldObj, aviInfra) {
					namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(aviInfra))
					key := lib.AviInfraSetting + "/" + utils.ObjKey(aviInfra)
					if err := validateAviInfraSetting(key, aviInfra); err != nil {
						utils.AviLog.Warnf("Error retrieved during validation of AviInfraSetting: %v", err)
					}
					utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				aviinfra, ok := obj.(*akov1alpha1.AviInfraSetting)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					aviinfra, ok = tombstone.Obj.(*akov1alpha1.AviInfraSetting)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an AviInfraSetting: %#v", obj)
						return
					}
				}
				key := lib.AviInfraSetting + "/" + utils.ObjKey(aviinfra)
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(aviinfra))
				utils.AviLog.Debugf("key: %s, msg: DELETE", key)
				// no need to validate for delete handler
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			},
		}

		informer.AviInfraSettingInformer.Informer().AddEventHandler(aviInfraEventHandler)
		informer.AviInfraSettingInformer.Informer().AddIndexers(
			cache.Indexers{
				lib.SeGroupAviSettingIndex: func(obj interface{}) ([]string, error) {
					infraSetting, ok := obj.(*akov1alpha1.AviInfraSetting)
					if !ok {
						return []string{}, nil
					}
					return []string{infraSetting.Spec.SeGroup.Name}, nil
				},
			},
		)
	}

	return
}

// SetupAKOCRDEventHandlers handles setting up of AKO CRD event handlers
func (c *AviController) SetupIstioCRDEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AKO Istio CRD Event handlers")
	informer := lib.GetIstioCRDInformers()

	virtualServiceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			vs := obj.(*istiov1alpha3.VirtualService)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
			key := lib.IstioVirtualService + "/" + utils.ObjKey(vs)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*istiov1alpha3.VirtualService)
			vs := new.(*istiov1alpha3.VirtualService)
			if !reflect.DeepEqual(oldObj.Spec, vs.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
				key := lib.IstioVirtualService + "/" + utils.ObjKey(vs)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
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
			c.workqueue[bkt].AddRateLimited(key)
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
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*istiov1alpha3.DestinationRule)
			dr := new.(*istiov1alpha3.DestinationRule)
			if !reflect.DeepEqual(oldObj.Spec, dr.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(dr))
				key := lib.IstioDestinationRule + "/" + utils.ObjKey(dr)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
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
			c.workqueue[bkt].AddRateLimited(key)
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
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if c.DisableSync {
				return
			}
			oldObj := old.(*istiov1alpha3.Gateway)
			vs := new.(*istiov1alpha3.Gateway)
			if !reflect.DeepEqual(oldObj.Spec, vs.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(vs))
				key := lib.IstioGateway + "/" + utils.ObjKey(vs)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
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
			c.workqueue[bkt].AddRateLimited(key)
		},
	}

	informer.GatewayInformer.Informer().AddEventHandler(gatewayEventHandler)

	return
}

// validateHostRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func validateHostRuleObj(key string, hostrule *akov1alpha1.HostRule) error {
	var err error
	fqdn := hostrule.Spec.VirtualHost.Fqdn
	foundHost, foundHR := objects.SharedCRDLister().GetFQDNToHostruleMapping(fqdn)
	if foundHost && foundHR != hostrule.Namespace+"/"+hostrule.Name {
		err = fmt.Errorf("duplicate fqdn %s found in %s", fqdn, foundHR)
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	refData := map[string]string{
		hostrule.Spec.VirtualHost.WAFPolicy:                  "WafPolicy",
		hostrule.Spec.VirtualHost.ApplicationProfile:         "AppProfile",
		hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name: "SslKeyCert",
		hostrule.Spec.VirtualHost.TLS.SSLProfile:             "SslProfile",
		hostrule.Spec.VirtualHost.AnalyticsProfile:           "AnalyticsProfile",
		hostrule.Spec.VirtualHost.ErrorPageProfile:           "ErrorPageProfile",
	}

	for _, policy := range hostrule.Spec.VirtualHost.HTTPPolicy.PolicySets {
		refData[policy] = "HttpPolicySet"
	}

	for _, script := range hostrule.Spec.VirtualHost.Datascripts {
		refData[script] = "VsDatascript"
	}

	if err := checkRefsOnController(key, refData); err != nil {
		status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	status.UpdateHostRuleStatus(key, hostrule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

func checkRefsOnController(key string, refMap map[string]string) error {
	for k, value := range refMap {
		if k == "" {
			continue
		}

		if err := checkRefOnController(key, value, k); err != nil {
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
	"ServiceEngineGroup":     "serviceenginegroup",
	"Network":                "network",
}

// checkRefOnController checks whether a provided ref on the controller
func checkRefOnController(key, refKey, refValue string) error {
	// assign the last avi client for ref checks
	aviClientLen := lib.GetshardSize()
	clients := avicache.SharedAVIClients()
	uri := fmt.Sprintf("/api/%s?name=%s&fields=name,type,labels,created_by", refModelMap[refKey], refValue)

	// For public clouds, check using network UUID in AWS, normal network API for GCP, skip altogether for Azure.
	if lib.IsPublicCloud() && refModelMap[refKey] == "network" {
		if lib.GetCloudType() == lib.CLOUD_AWS {
			var rest_response interface{}
			utils.AviLog.Infof("Cloud is AWS, checking network ref using uuid")
			uri := fmt.Sprintf("/api/%s/%s", refModelMap[refKey], refValue)
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
		} else if lib.GetCloudType() == lib.CLOUD_AZURE {
			utils.AviLog.Infof("key: %s, msg: Skipping network name check for Azure Cloud", key)
			return nil
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
		if appProfType, ok := item["type"].(string); ok && appProfType != lib.AllowedApplicationProfile {
			utils.AviLog.Warnf("key: %s, msg: applicationProfile: %s must be of type %s", key, refValue, lib.AllowedApplicationProfile)
			return fmt.Errorf("%s \"%s\" found on controller is invalid, must be of type: %s",
				refModelMap[refKey], refValue, lib.AllowedApplicationProfile)
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

// validateHTTPRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func validateHTTPRuleObj(key string, httprule *akov1alpha1.HTTPRule) error {
	refData := make(map[string]string)
	for _, path := range httprule.Spec.Paths {
		refData[path.TLS.SSLProfile] = "SslProfile"
		refData[path.ApplicationPersistence] = "ApplicationPersistence"

		for _, hm := range path.HealthMonitors {
			refData[hm] = "HealthMonitor"
		}
	}

	if err := checkRefsOnController(key, refData); err != nil {
		status.UpdateHTTPRuleStatus(key, httprule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	status.UpdateHTTPRuleStatus(key, httprule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

// validateAviInfraSetting would do validaion checks on the
// ingested AviInfraSetting objects
func validateAviInfraSetting(key string, infraSetting *akov1alpha1.AviInfraSetting) error {
	if ((infraSetting.Spec.Network.EnableRhi != nil && !*infraSetting.Spec.Network.EnableRhi) || infraSetting.Spec.Network.EnableRhi == nil) &&
		len(infraSetting.Spec.Network.BgpPeerLabels) > 0 {
		err := fmt.Errorf("BGPPeerLabels cannot be set if EnableRhi is false.")
		status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	refData := make(map[string]string)
	for _, vipNetwork := range infraSetting.Spec.Network.VipNetworks {
		if vipNetwork.Cidr != "" {
			re := regexp.MustCompile(lib.IPCIDRRegex)
			if !re.MatchString(vipNetwork.Cidr) {
				err := fmt.Errorf("invalid CIDR configuration %s detected for networkName %s in vipNetworkList", vipNetwork.Cidr, vipNetwork.NetworkName)
				status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
					Status: lib.StatusRejected,
					Error:  err.Error(),
				})
				return err
			}
		}
		refData[vipNetwork.NetworkName] = "Network"
	}

	if infraSetting.Spec.SeGroup.Name != "" {
		refData[infraSetting.Spec.SeGroup.Name] = "ServiceEngineGroup"
	}

	if err := checkRefsOnController(key, refData); err != nil {
		status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		return err
	}

	// This would add SEG labels only if they are not configured yet. In case there is a label mismatch
	// to any pre-existing SEG labels, the AviInfraSettig CR will get Rejected from the checkRefsOnController
	// step before this.
	if infraSetting.Spec.SeGroup.Name != "" {
		addSeGroupLabel(key, infraSetting.Spec.SeGroup.Name)
	}

	status.UpdateAviInfraSettingStatus(key, infraSetting, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

// addSeGroupLabel configures SEGroup with appropriate labels, during AviInfraSetting
// creation/updates after ingestion
func addSeGroupLabel(key, segName string) {
	// assign the last avi client for ref checks
	clients := avicache.SharedAVIClients()
	aviClientLen := lib.GetshardSize()

	// configure labels on SeGroup if not present already.
	seGroup, err := avicache.GetAviSeGroup(clients.AviClient[aviClientLen], segName)
	if err != nil {
		utils.AviLog.Error(err)
		return
	}

	avicache.ConfigureSeGroupLabels(clients.AviClient[aviClientLen], seGroup)
}

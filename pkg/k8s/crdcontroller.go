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
	"strings"
	"time"

	akov1alpha1 "ako/pkg/apis/ako/v1alpha1"
	avicache "ako/pkg/cache"
	akocrd "ako/pkg/client/clientset/versioned"
	akoinformers "ako/pkg/client/informers/externalversions"
	"ako/pkg/lib"
	"ako/pkg/objects"
	"ako/pkg/status"

	"github.com/avinetworks/container-lib/utils"
	"github.com/avinetworks/sdk/go/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func NewCRDInformers(cs akocrd.Interface) {
	var akoInformerFactory akoinformers.SharedInformerFactory

	akoInformerFactory = akoinformers.NewSharedInformerFactoryWithOptions(cs, time.Second*30)
	hostRuleInformer := akoInformerFactory.Ako().V1alpha1().HostRules()
	httpRuleInformer := akoInformerFactory.Ako().V1alpha1().HTTPRules()

	lib.SetCRDInformers(&lib.AKOCrdInformers{
		HostRuleInformer: hostRuleInformer,
		HTTPRuleInformer: httpRuleInformer,
	})
}

// SetupAKOCRDEventHandlers handles setting up of AKO CRD event handlers
func (c *AviController) SetupAKOCRDEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AKO CRD Event handlers")
	informer := lib.GetCRDInformers()

	hostRuleEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for hostrule update")
				return
			}
			hostrule := obj.(*akov1alpha1.HostRule)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
			key := lib.HostRule + "/" + utils.ObjKey(hostrule)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			if err := validateHostRuleObj(key, hostrule); err == nil {
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj := old.(*akov1alpha1.HostRule)
			hostrule := new.(*akov1alpha1.HostRule)
			if !reflect.DeepEqual(oldObj.Spec, hostrule.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
				key := lib.HostRule + "/" + utils.ObjKey(hostrule)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				if err := validateHostRuleObj(key, hostrule); err == nil {
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for hostrule update")
				return
			}
			hostrule := obj.(*akov1alpha1.HostRule)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
			key := lib.HostRule + "/" + utils.ObjKey(hostrule)
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
	}

	httpRuleEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for httprule update")
				return
			}
			httprule := obj.(*akov1alpha1.HTTPRule)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
			key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			if err := validateHTTPRuleObj(key, httprule); err == nil {
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj := old.(*akov1alpha1.HTTPRule)
			httprule := new.(*akov1alpha1.HTTPRule)
			if !reflect.DeepEqual(oldObj.Spec, httprule.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
				key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				if err := validateHTTPRuleObj(key, httprule); err == nil {
					bkt := utils.Bkt(namespace, numWorkers)
					c.workqueue[bkt].AddRateLimited(key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				utils.AviLog.Debugf("Sync disabled, skipping sync for httprule update")
				return
			}
			httprule := obj.(*akov1alpha1.HTTPRule)
			key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
			utils.AviLog.Debugf("key: %s, msg: DELETE", key)
			// no need to validate for delete handler
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
	}

	informer.HostRuleInformer.Informer().AddEventHandler(hostRuleEventHandler)
	informer.HTTPRuleInformer.Informer().AddEventHandler(httpRuleEventHandler)

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
		status.UpdateHostRuleStatus(hostrule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		utils.AviLog.Errorf("key: %s, msg: %v", key, err)
		return err
	}

	refData := map[string]string{
		hostrule.Spec.VirtualHost.WAFPolicy:                  "WafPolicy",
		hostrule.Spec.VirtualHost.NetworkSecurityPolicy:      "NsPolicy",
		hostrule.Spec.VirtualHost.ApplicationProfile:         "AppProfile",
		hostrule.Spec.VirtualHost.TLS.SSLKeyCertificate.Name: "SslKeyCert",
	}
	for _, policy := range hostrule.Spec.VirtualHost.HTTPPolicy.PolicySets {
		refData[policy] = "HttpPolicySet"
	}

	// TODO (shchauhan): optimisation opportunity to make batched api calls
	// or distribute in threads
	for k, value := range refData {
		if k == "" {
			continue
		}

		if !checkRefOnController(value, k) {
			err = fmt.Errorf("%s \"%s\" not found on controller", value, k)
			status.UpdateHostRuleStatus(hostrule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  err.Error(),
			})
			utils.AviLog.Errorf("key: %s, msg: %v", key, err)
			return err
		}
	}
	status.UpdateHostRuleStatus(hostrule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

// validateHTTPRuleObj would do validation checks
// update internal CRD caches, and push relevant ingresses to ingestion
func validateHTTPRuleObj(key string, httprule *akov1alpha1.HTTPRule) error {
	var err error
	refData := make(map[string]string)
	for _, path := range httprule.Spec.Paths {
		refData[path.TLS.SSLProfile] = "SslProfile"
	}

	for k, value := range refData {
		if k == "" {
			continue
		}

		if !checkRefOnController(value, k) {
			err = fmt.Errorf("%s \"%s\" not found on controller or is invalid", value, k)
			status.UpdateHTTPRuleStatus(httprule, status.UpdateCRDStatusOptions{
				Status: lib.StatusRejected,
				Error:  err.Error(),
			})
			utils.AviLog.Errorf("key: %s, msg: %v", key, err)
			return err
		}
	}

	hostrule := httprule.Spec.HostRule
	hostruleNSName := strings.Split(hostrule, "/")
	_, err = lib.GetCRDClientset().AkoV1alpha1().HostRules(hostruleNSName[0]).Get(hostruleNSName[1], metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("hostrules.ako.k8s.io %s not found or is invalid", hostrule)
		status.UpdateHTTPRuleStatus(httprule, status.UpdateCRDStatusOptions{
			Status: lib.StatusRejected,
			Error:  err.Error(),
		})
		utils.AviLog.Error(err)
		return nil
	}

	status.UpdateHTTPRuleStatus(httprule, status.UpdateCRDStatusOptions{
		Status: lib.StatusAccepted,
		Error:  "",
	})
	return nil
}

var refModelMap = map[string]string{
	"SslKeyCert":    "sslkeyandcertificate",
	"WafPolicy":     "wafpolicy",
	"NsPolicy":      "networksecuritypolicy",
	"HttpPolicySet": "httppolicyset",
	"SslProfile":    "sslprofile",
	"AppProfile":    "applicationprofile",
}

// checkRefOnController checks whether a provided ref on the controller
func checkRefOnController(refKey, refValue string) bool {
	uri := fmt.Sprintf("/api/%s?name=%s&fields=name,type", refModelMap[refKey], refValue)
	clients := avicache.SharedAVIClients()

	// assign the last avi client for ref checks
	aviClientLen := lib.GetshardSize()
	result, err := avicache.AviGetCollectionRaw(clients.AviClient[aviClientLen], uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return false
	}

	if result.Count == 0 {
		utils.AviLog.Warnf("No Objects found for refName: %s/%s", refModelMap[refKey], refValue)
		return false
	}

	if refKey == "AppProfile" {
		items := make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &items)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			return false
		}

		appProf := models.ApplicationProfile{}
		err := json.Unmarshal(items[0], &appProf)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			return false
		}

		if *appProf.Type != lib.AllowedApplicationProfile {
			utils.AviLog.Warnf("applicationProfile: %s must be of type %s", refValue, lib.AllowedApplicationProfile)
			return false
		}
	}

	utils.AviLog.Infof("Ref found for %s/%s", refModelMap[refKey], refValue)
	return true
}

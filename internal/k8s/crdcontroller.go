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
	"reflect"
	"time"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/apis/ako/v1alpha1"
	akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/clientset/versioned"
	akoinformers "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/informers/externalversions"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

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

// SetupAKOCRDEventHandlers handles setting up of AKO CRD event handlers
func (c *AviController) SetupAKOCRDEventHandlers(numWorkers uint32) {
	utils.AviLog.Infof("Setting up AKO CRD Event handlers")
	informer := lib.GetCRDInformers()

	hostRuleEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			hostrule := obj.(*akov1alpha1.HostRule)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
			key := lib.HostRule + "/" + utils.ObjKey(hostrule)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj := old.(*akov1alpha1.HostRule)
			hostrule := new.(*akov1alpha1.HostRule)
			if !reflect.DeepEqual(oldObj.Spec, hostrule.Spec) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(hostrule))
				key := lib.HostRule + "/" + utils.ObjKey(hostrule)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
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
				return
			}
			httprule := obj.(*akov1alpha1.HTTPRule)
			namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
			key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
			utils.AviLog.Debugf("key: %s, msg: ADD", key)
			bkt := utils.Bkt(namespace, numWorkers)
			c.workqueue[bkt].AddRateLimited(key)
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj := old.(*akov1alpha1.HTTPRule)
			httprule := new.(*akov1alpha1.HTTPRule)
			// reflect.DeepEqual does not work on type []byte,
			// unable to capture edits in destinationCA
			if isHTTPRuleUpdated(oldObj, httprule) {
				namespace, _, _ := cache.SplitMetaNamespaceKey(utils.ObjKey(httprule))
				key := lib.HTTPRule + "/" + utils.ObjKey(httprule)
				utils.AviLog.Debugf("key: %s, msg: UPDATE", key)
				bkt := utils.Bkt(namespace, numWorkers)
				c.workqueue[bkt].AddRateLimited(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
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

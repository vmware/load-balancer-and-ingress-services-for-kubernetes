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
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

func (c *GatewayController) SetupCRDEventHandlers(numWorkers uint32) {
	if !utils.IsWCP() {
		L7CRDEventHandler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				_, ok := obj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Warn("Error in converting object to L7 CRD object")
					return
				}
				// fetch name and namespace of appprofile crd
				namespace, name := getNamespaceName(obj)
				if namespace == "" || name == "" {
					return
				}
				isProcessed, _ := isObjectProcessed(obj, namespace, name)
				if !isProcessed {
					return
				}
				key := lib.L7Rule + "/" + namespace + "/" + name
				c.processHTTPRoutes(key, namespace, name, numWorkers)
			},
			DeleteFunc: func(obj interface{}) {
				if c.DisableSync {
					return
				}
				_, ok := obj.(*unstructured.Unstructured)
				if !ok {
					// httpRoute was deleted but its final state is unrecorded.
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
						return
					}
					_, ok = tombstone.Obj.(*unstructured.Unstructured)
					if !ok {
						utils.AviLog.Errorf("Tombstone contained object that is not an L7Rule: %#v", obj)
						return
					}
				}
				// fetch name and namespace of appprofile crd
				namespace, name := getNamespaceName(obj)
				if namespace == "" || name == "" {
					return
				}
				isProcessed, _ := isObjectProcessed(obj, namespace, name)
				if !isProcessed {
					return
				}
				key := lib.L7Rule + "/" + namespace + "/" + name
				// process HTTP Route to remove L7Rule settings.
				c.processHTTPRoutes(key, namespace, name, numWorkers)
			},
			UpdateFunc: func(old, cur interface{}) {
				if c.DisableSync {
					return
				}
				namespace, name := getNamespaceName(old)
				if namespace == "" || name == "" {
					return
				}
				isOldObjProcessed, _ := isObjectProcessed(old, namespace, name)
				isCurObjProcessed, _ := isObjectProcessed(cur, namespace, name)

				if !isOldObjProcessed && !isCurObjProcessed {
					utils.AviLog.Warnf("key: %s/%s, msg: L7Rule is not processed.", namespace, name)
					return
				}

				key := lib.L7Rule + "/" + namespace + "/" + name
				c.processHTTPRoutes(key, namespace, name, numWorkers)

			},
		}
		c.dynamicInformers.L7CRDInformer.Informer().AddEventHandler(L7CRDEventHandler)
	}
}

func getNamespaceName(obj interface{}) (string, string) {
	oldObj := obj.(*unstructured.Unstructured)
	name := oldObj.GetName()
	namespace := oldObj.GetNamespace()
	return namespace, name
}
func isObjectProcessed(obj interface{}, namespace, name string) (bool, string) {
	oldObj := obj.(*unstructured.Unstructured)
	statusJSON, found, err := unstructured.NestedMap(oldObj.UnstructuredContent(), "status")
	if err != nil || !found {
		utils.AviLog.Warnf("key:%s/%s, msg:ApplicationProfile CRD status not found: %+v", namespace, name, err)
		return false, ""
	}
	// fetch the status
	status, ok := statusJSON["status"]
	if !ok || status == "" {
		utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD status not found", namespace, name)
		return false, ""
	}
	return true, status.(string)
}

func (c *GatewayController) processHTTPRoutes(key, namespace, name string, numWorkers uint32) {
	utils.AviLog.Debugf("key: %s, msg: Fetchting HTTPRoute associated with L7Rule %s/%s", key, namespace, name)

	l7RuleNSName := namespace + "/" + name
	ok, httpRoutes := akogatewayapiobjects.GatewayApiLister().GetL7RuleToHTTPRouteMapping(l7RuleNSName)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: No HTTPRoute associated with L7Rule", key)
		return
	}
	for httpRoute := range httpRoutes {
		utils.AviLog.Debugf("key: %s, Processing HTTPRoute %s", key, httpRoute)
		namespace, name, _ = cache.SplitMetaNamespaceKey(httpRoute)
		// Get HTTPRoute-->L7Rule mapping
		_, l7RuleNSNameList := akogatewayapiobjects.GatewayApiLister().GetHTTPRouteToL7RuleMapping(httpRoute)
		ok := l7RuleNSNameList[l7RuleNSName]
		if !ok {
			// IF HTTPRoute to L7Rule Mapping is no there.. delete the entry from L7Rule to HTTPRoute
			akogatewayapiobjects.GatewayApiLister().DeleteL7RuleToHTTPRouteMapping(l7RuleNSName, httpRoute)
			continue
		}

		// fetch httpRoute
		httpRouteObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				// delete mapping for L7Rule--->HTTPRoute
				akogatewayapiobjects.GatewayApiLister().DeleteL7RuleToHTTPRouteMapping(l7RuleNSName, httpRoute)
			}
			utils.AviLog.Warnf("key: %s, msg: Error while fetching HTTPRoute object. Error: %+v ", key, err)
			continue
		}

		if !IsHTTPRouteConfigValid(key, httpRouteObj) {
			continue
		}
		bkt := utils.Bkt(namespace, numWorkers)
		httpRouteKey := lib.HTTPRoute + "/" + namespace + "/" + name
		utils.AviLog.Debugf("key: %s, msg: HTTPRoute add: %s", key, httpRouteKey)
		c.workqueue[bkt].AddRateLimited(httpRouteKey)
	}
}

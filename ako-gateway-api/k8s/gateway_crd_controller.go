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
	"strings"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

func (c *GatewayController) SetupCRDEventHandlers(numWorkers uint32) {
	// To be removed once the CRDs are merged in Supervisor
	if utils.IsWCP() {
		return
	}
	if !utils.IsWCP() {
		c.setupL7CRDEventHandlers(numWorkers)
	}
	// Skip setup if AKO CRD Operator is not enabled
	if !lib.IsAKOCRDOperatorEnabled() {
		utils.AviLog.Warnf("Skipping event handler setup for AKO CRD Operator managed CRDs as it is not enabled")
		return
	}
	c.setupHealthMonitorEventHandlers(numWorkers)
	c.setupRouteBackendExtensionEventHandler(numWorkers)
	c.setupApplicationProfileEventHandlers(numWorkers)
}

func (c *GatewayController) setupL7CRDEventHandlers(numWorkers uint32) {
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
			l7RuleObj, ok := obj.(*unstructured.Unstructured)
			if !ok {
				// httpRoute was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				l7RuleObj, ok = tombstone.Obj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an L7Rule: %#v", obj)
					return
				}
			}
			// fetch name and namespace of appprofile crd
			namespace, name := getNamespaceName(l7RuleObj)
			if namespace == "" || name == "" {
				return
			}
			isProcessed, _ := isObjectProcessed(l7RuleObj, namespace, name)
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

func (c *GatewayController) setupHealthMonitorEventHandlers(numWorkers uint32) {
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

			key := akogatewayapilib.HealthMonitorKind + "/" + namespace + "/" + name
			processed, _, err := lib.IsHealthMonitorProcessedWithOptions(key, namespace, name, akogatewayapilib.GetDynamicClientSet(), false, healthMonitorObj)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: error: Error processing HealthMonitor. err: %s", key, err)
				return
			}
			if !processed {
				utils.AviLog.Warnf("key: %s, msg: HealthMonitor is not processed by ako-crd-operator. err: %s", key, err)
				return
			}
			c.processHTTPRoutesForHealthMonitor(key, namespace, name, numWorkers)
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
			key := akogatewayapilib.HealthMonitorKind + "/" + namespace + "/" + name
			processed, _, err := lib.IsHealthMonitorProcessedWithOptions(key, namespace, name, akogatewayapilib.GetDynamicClientSet(), false, healthMonitorObj)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: error: Error processing HealthMonitor. err: %s", key, err)
				return
			}
			if !processed {
				utils.AviLog.Warnf("key: %s, msg: HealthMonitor is not processed by ako-crd-operator. err: %s", key, err)
				return
			}
			// process HTTP Route to remove HealthMonitor mappings
			c.processHTTPRoutesForHealthMonitor(key, namespace, name, numWorkers)
		},
		UpdateFunc: func(oldObj, curObj interface{}) {
			if c.DisableSync {
				return
			}
			oldHealthMonitorObj, ok := oldObj.(*unstructured.Unstructured)
			if !ok {
				utils.AviLog.Warn("Error in converting object to HealthMonitor object")
				return
			}

			curHealthMonitorObj, ok := curObj.(*unstructured.Unstructured)
			if !ok {
				utils.AviLog.Warn("Error in converting object to HealthMonitor object")
				return
			}
			namespace, name := oldHealthMonitorObj.GetNamespace(), oldHealthMonitorObj.GetName()
			if namespace == "" || name == "" {
				return
			}

			namespace, name = curHealthMonitorObj.GetNamespace(), curHealthMonitorObj.GetName()
			if namespace == "" || name == "" {
				return
			}

			key := akogatewayapilib.HealthMonitorKind + "/" + namespace + "/" + name
			processedOldObj, _, _ := lib.IsHealthMonitorProcessedWithOptions(key, namespace, name, akogatewayapilib.GetDynamicClientSet(), false, oldHealthMonitorObj)
			processedCurObj, _, _ := lib.IsHealthMonitorProcessedWithOptions(key, namespace, name, akogatewayapilib.GetDynamicClientSet(), false, curHealthMonitorObj)

			if !processedOldObj && !processedCurObj {
				utils.AviLog.Debugf("key: %s/%s, msg: HealthMonitor is not processed.", namespace, name)
				return
			}

			c.processHTTPRoutesForHealthMonitor(key, namespace, name, numWorkers)
		},
	}

	c.dynamicInformers.HealthMonitorInformer.Informer().AddEventHandler(healthMonitorEventHandler)
}

func (c *GatewayController) processHTTPRoutesForHealthMonitor(key, namespace, name string, numWorkers uint32) {
	utils.AviLog.Debugf("key: %s, msg: Fetching HTTPRoute associated with HealthMonitor %s/%s", key, namespace, name)
	hmNameNS := namespace + "/" + name
	ok, httpRoutes := akogatewayapiobjects.GatewayApiLister().GetHealthMonitorToHTTPRoutesMapping(hmNameNS)
	if !ok {
		utils.AviLog.Debugf("key: %s, msg: No HTTPRoute associated with HealthMonitor", key)
		return
	}
	for httpRoute := range httpRoutes {
		utils.AviLog.Debugf("key: %s, Processing HTTPRoute %s", key, httpRoute)
		httpRouteNamespace, httpRouteName, _ := cache.SplitMetaNamespaceKey(httpRoute)
		// Get HTTPRoute-->Healthmonitor mapping
		_, healthMonitorNSNameList := akogatewayapiobjects.GatewayApiLister().GetHTTPRouteToHealthMonitorMapping(httpRoute)
		_, ok := healthMonitorNSNameList[hmNameNS]
		if !ok {
			// delete mapping for HealthMonitor--->HTTPRoute
			akogatewayapiobjects.GatewayApiLister().DeleteHealthMonitorToHTTPRoutesMapping(hmNameNS, httpRoute)
			continue
		}

		// fetch httpRoute
		httpRouteObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(httpRouteNamespace).Get(httpRouteName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				// delete mapping for HealthMonitor--->HTTPRoute
				akogatewayapiobjects.GatewayApiLister().DeleteHealthMonitorToHTTPRoutesMapping(hmNameNS, httpRoute)
			}
			utils.AviLog.Warnf("key: %s, msg: Error while fetching HTTPRoute object. Error: %+v ", key, err)
			continue
		}

		if !IsHTTPRouteConfigValid(key, httpRouteObj) {
			continue
		}
		bkt := utils.Bkt(httpRouteNamespace, numWorkers)
		httpRouteKey := lib.HTTPRoute + "/" + httpRouteNamespace + "/" + httpRouteName
		utils.AviLog.Debugf("key: %s, msg: HTTPRoute add: %s", key, httpRouteKey)
		c.workqueue[bkt].AddRateLimited(httpRouteKey)
	}
}

func getNamespaceName(obj interface{}) (string, string) {
	unstructuredObj := obj.(*unstructured.Unstructured)
	name := unstructuredObj.GetName()
	namespace := unstructuredObj.GetNamespace()
	return namespace, name
}

func isObjectProcessed(obj interface{}, namespace, name string) (bool, string) {
	crdObj := obj.(*unstructured.Unstructured)
	statusJSON, found, err := unstructured.NestedMap(crdObj.UnstructuredContent(), "status")
	if err != nil || !found {
		utils.AviLog.Warnf("key:%s/%s, msg: L7Rule CRD status not found: %+v", namespace, name, err)
		return false, ""
	}
	// fetch the status
	status, ok := statusJSON["status"]
	if !ok || status == "" {
		utils.AviLog.Warnf("key:%s/%s, msg: L7Rule  CRD status not found", namespace, name)
		return false, ""
	}
	return true, status.(string)
}

func (c *GatewayController) processHTTPRoutes(key, namespace, name string, numWorkers uint32) {
	keySplit := strings.Split(key, "/")
	if len(keySplit) != 3 {
		utils.AviLog.Errorf("key:%s, msg: Unable to parse CR key", key)
		return
	}
	crType := keySplit[0]
	switch crType {
	case lib.L7Rule:
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
			_, ok := l7RuleNSNameList[l7RuleNSName]
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
	case akogatewayapilib.RouteBackendExtensionKind:
		utils.AviLog.Debugf("key: %s, msg: Fetchting HTTPRoute associated with RouteBackendExtension %s/%s", key, namespace, name)

		routeBackendExtensionNSName := namespace + "/" + name
		ok, httpRoutes := akogatewayapiobjects.GatewayApiLister().GetRouteBackendExtensionToHTTPRouteMapping(routeBackendExtensionNSName)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: No HTTPRoute associated with RouteBackendExtension", key)
			return
		}
		for httpRoute := range httpRoutes {
			utils.AviLog.Debugf("key: %s, Processing HTTPRoute %s", key, httpRoute)
			namespace, name, _ = cache.SplitMetaNamespaceKey(httpRoute)
			// Get HTTPRoute-->RouteBackendExtension mapping
			_, routeBackendExtensionNSNameList := akogatewayapiobjects.GatewayApiLister().GetHTTPRouteToRouteBackendExtensionMapping(httpRoute)
			_, ok := routeBackendExtensionNSNameList[routeBackendExtensionNSName]
			if !ok {
				// If HTTPRoute to RouteBackendExtension Mapping is not there, delete the entry from RouteBackendExtension to HTTPRoute
				akogatewayapiobjects.GatewayApiLister().DeleteRouteBackendExtensionToHTTPRouteMapping(routeBackendExtensionNSName, httpRoute)
				continue
			}

			// fetch httpRoute
			httpRouteObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					// delete mapping for RouteBackendExtension--->HTTPRoute
					akogatewayapiobjects.GatewayApiLister().DeleteRouteBackendExtensionToHTTPRouteMapping(routeBackendExtensionNSName, httpRoute)
				}
				utils.AviLog.Warnf("key: %s, msg: Error while fetching HTTPRoute object : %+v ", key, err)
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
}

func (c *GatewayController) setupApplicationProfileEventHandlers(numWorkers uint32) {
	appProfileEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}

			_, ok := obj.(*unstructured.Unstructured)
			if !ok {
				utils.AviLog.Warn("Error in converting object to ApplicationProfile CRD object")
				return
			}

			// fetch name and namespace of app profile crd
			namespace, name := getNamespaceName(obj)
			if namespace == "" || name == "" {
				return
			}
			isProcessed := akogatewayapilib.IsApplicationProfileProcessed(obj, namespace, name)
			if !isProcessed {
				return
			}
			key := lib.ApplicationProfile + "/" + namespace + "/" + name
			c.processApplicationProfiles(key, namespace, name, numWorkers)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			applicationProfileObj, ok := obj.(*unstructured.Unstructured)
			if !ok {
				// httpRoute was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				applicationProfileObj, ok = tombstone.Obj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not an ApplicationProfile: %#v", obj)
					return
				}
			}
			// fetch name and namespace of appprofile crd
			namespace, name := getNamespaceName(applicationProfileObj)
			if namespace == "" || name == "" {
				return
			}
			isProcessed := akogatewayapilib.IsApplicationProfileProcessed(applicationProfileObj, namespace, name)
			if !isProcessed {
				return
			}
			key := lib.ApplicationProfile + "/" + namespace + "/" + name
			// process HTTP Route to remove ApplicationProfile settings.
			c.processApplicationProfiles(key, namespace, name, numWorkers)
		},
		UpdateFunc: func(old, cur interface{}) {
			if c.DisableSync {
				return
			}
			namespace, name := getNamespaceName(old)
			if namespace == "" || name == "" {
				return
			}
			isOldObjProcessed := akogatewayapilib.IsApplicationProfileProcessed(old, namespace, name)
			isCurObjProcessed := akogatewayapilib.IsApplicationProfileProcessed(cur, namespace, name)

			if !isOldObjProcessed && !isCurObjProcessed {
				return
			}

			key := lib.ApplicationProfile + "/" + namespace + "/" + name
			c.processApplicationProfiles(key, namespace, name, numWorkers)
		},
	}
	c.dynamicInformers.AppProfileCRDInformer.Informer().AddEventHandler(appProfileEventHandler)
}

func (c *GatewayController) processApplicationProfiles(key, namespace, name string, numWorkers uint32) {
	utils.AviLog.Debugf("key: %s, msg: Fetchting HTTPRoutes associated with ApplicationProfile %s/%s", key, namespace, name)

	appProfileNsName := namespace + "/" + name
	ok, httpRoutes := akogatewayapiobjects.GatewayApiLister().GetApplicationProfileToHTTPRouteMapping(appProfileNsName)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: No HTTPRoute associated with ApplicationProfile %s", key, appProfileNsName)
		return
	}

	for httpRoute := range httpRoutes {
		utils.AviLog.Debugf("key: %s, Processing HTTPRoute %s", key, httpRoute)

		namespace, name, _ = cache.SplitMetaNamespaceKey(httpRoute)
		// Get application profile names from HTTPRoute --> ApplicationProfile mapping
		_, appProfileNSNameList := akogatewayapiobjects.GatewayApiLister().GetHTTPRouteToApplicationProfileMapping(httpRoute)
		_, ok := appProfileNSNameList[appProfileNsName]
		if !ok {
			// If HTTPRoute to ApplicationProfile mapping is not present, delete the entry from ApplicationProfile --> HTTPRoute map
			akogatewayapiobjects.GatewayApiLister().DeleteApplicationProfileToHTTPRouteMapping(appProfileNsName, httpRoute)
			continue
		}

		// fetch httpRoute
		httpRouteObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				// delete mapping for ApplicationProfile --> HTTPRoute
				akogatewayapiobjects.GatewayApiLister().DeleteApplicationProfileToHTTPRouteMapping(appProfileNsName, httpRoute)
			}
			utils.AviLog.Warnf("key: %s, msg: Error while fetching HTTPRoute object. Error: %+v ", key, err)
			continue
		}

		if !IsHTTPRouteConfigValid(key, httpRouteObj) {
			continue
		}

		bkt := utils.Bkt(namespace, numWorkers)
		httpRouteKey := lib.HTTPRoute + "/" + namespace + "/" + name
		utils.AviLog.Warnf("key: %s, msg: HTTPRoute add: %s", key, httpRouteKey)
		c.workqueue[bkt].AddRateLimited(httpRouteKey)
	}
}

func (c *GatewayController) setupRouteBackendExtensionEventHandler(numWorkers uint32) {
	RouteBackendExtensionCRDEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			object, ok := obj.(*unstructured.Unstructured)
			if !ok {
				utils.AviLog.Warn("Error in converting object to RouteBackendExtension CRD object")
				return
			}
			// fetch name and namespace of RouteBackendExtension crd
			namespace, name := getNamespaceName(object)
			if namespace == "" || name == "" {
				return
			}
			key := akogatewayapilib.RouteBackendExtensionKind + "/" + namespace + "/" + name
			isProcessed, _, _ := akogatewayapilib.IsRouteBackendExtensionProcessed(key, namespace, name, object)
			if !isProcessed {
				return
			}
			c.processHTTPRoutes(key, namespace, name, numWorkers)
		},
		DeleteFunc: func(obj interface{}) {
			if c.DisableSync {
				return
			}
			object, ok := obj.(*unstructured.Unstructured)
			if !ok {
				// httpRoute was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				object, ok = tombstone.Obj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Errorf("Tombstone contained object that is not a RouteBackendExtension: %#v", obj)
					return
				}
			}
			// fetch name and namespace of RouteBackendExtension crd
			namespace, name := getNamespaceName(object)
			if namespace == "" || name == "" {
				return
			}
			key := akogatewayapilib.RouteBackendExtensionKind + "/" + namespace + "/" + name
			isProcessed, _, _ := akogatewayapilib.IsRouteBackendExtensionProcessed(key, namespace, name, object)
			if !isProcessed {
				return
			}
			// process HTTP Route to remove RouteBackendExtension settings.
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
			key := akogatewayapilib.RouteBackendExtensionKind + "/" + namespace + "/" + name
			oldObj, ok := old.(*unstructured.Unstructured)
			if !ok {
				utils.AviLog.Warn("Error in converting old object to RouteBackendExtension CRD object")
				return
			}
			curObj, ok := cur.(*unstructured.Unstructured)
			if !ok {
				utils.AviLog.Warn("Error in converting current object to RouteBackendExtension CRD object")
				return
			}
			isOldObjProcessed, _, _ := akogatewayapilib.IsRouteBackendExtensionProcessed(key, namespace, name, oldObj)
			isCurObjProcessed, _, _ := akogatewayapilib.IsRouteBackendExtensionProcessed(key, namespace, name, curObj)

			if !isOldObjProcessed && !isCurObjProcessed {
				utils.AviLog.Warnf("key: %s, msg: RouteBackendExtension is not processed", key)
				return
			}

			c.processHTTPRoutes(key, namespace, name, numWorkers)

		},
	}
	c.dynamicInformers.RouteBackendExtensionCRDInformer.Informer().AddEventHandler(RouteBackendExtensionCRDEventHandler)
}

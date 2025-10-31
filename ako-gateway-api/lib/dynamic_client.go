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

package lib

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var dynamicInformerInstance *DynamicInformers
var dynamicClientSet dynamic.Interface

var (
	L7CRDGVR = schema.GroupVersionResource{
		Group:    "ako.vmware.com",
		Version:  "v1alpha2",
		Resource: "l7rules",
	}
	HealthMonitorGVR = schema.GroupVersionResource{
		Group:    "ako.vmware.com",
		Version:  "v1alpha1",
		Resource: "healthmonitors",
	}
	RouteBackendExtensionCRDGVR = schema.GroupVersionResource{
		Group:    "ako.vmware.com",
		Version:  "v1alpha1",
		Resource: "routebackendextensions",
	}
	AppProfileCRDGVR = schema.GroupVersionResource{
		Group:    "ako.vmware.com",
		Version:  "v1alpha1",
		Resource: "applicationprofiles",
	}
	PKIProfileCRDGVR = schema.GroupVersionResource{
		Group:    "ako.vmware.com",
		Version:  "v1alpha1",
		Resource: "pkiprofiles",
	}
)

// NewDynamicClientSet initializes dynamic client set instance
func NewDynamicClientSet(config *rest.Config) (dynamic.Interface, error) {

	ds, err := dynamic.NewForConfig(config)
	if err != nil {
		utils.AviLog.Infof("Error while creating dynamic client %v", err)
		return nil, err
	}
	if dynamicClientSet == nil {
		dynamicClientSet = ds
	}
	return dynamicClientSet, nil
}

// SetDynamicClientSet is used for Unit tests.
func SetDynamicClientSet(c dynamic.Interface) {
	dynamicClientSet = c
}

// GetDynamicClientSet returns dynamic client set instance
func GetDynamicClientSet() dynamic.Interface {
	if dynamicClientSet == nil {
		utils.AviLog.Warn("Cannot retrieve the dynamic clientset since it's not initialized yet.")
		return nil
	}
	return dynamicClientSet
}

// DynamicInformers holds third party generic informers
type DynamicInformers struct {
	L7CRDInformer                    informers.GenericInformer
	HealthMonitorInformer            informers.GenericInformer
	RouteBackendExtensionCRDInformer informers.GenericInformer
	AppProfileCRDInformer            informers.GenericInformer
}

// NewDynamicInformers initializes the DynamicInformers struct
func NewDynamicInformers(client dynamic.Interface, akoInfra bool) *DynamicInformers {
	informers := &DynamicInformers{}
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)

	// not applicable in wcp context
	if !utils.IsWCP() {
		informers.L7CRDInformer = f.ForResource(L7CRDGVR)
	}
	// Initialize HealthMonitor, ApplicationProfile and RouteBackendExtension informers only when AKO CRD Operator is enabled
	if lib.IsAKOCRDOperatorEnabled() {
		informers.HealthMonitorInformer = f.ForResource(HealthMonitorGVR)
		informers.AppProfileCRDInformer = f.ForResource(AppProfileCRDGVR)
		informers.RouteBackendExtensionCRDInformer = f.ForResource(RouteBackendExtensionCRDGVR)
	}
	dynamicInformerInstance = informers
	return dynamicInformerInstance
}

// GetDynamicInformers returns DynamicInformers instance
func GetDynamicInformers() *DynamicInformers {
	if dynamicInformerInstance == nil {
		utils.AviLog.Warn("Cannot retrieve the dynamic informers since it's not initialized yet.")
		return nil
	}
	return dynamicInformerInstance
}

func IsRouteBackendExtensionProcessed(key, namespace, name string, objects ...*unstructured.Unstructured) (bool, string, error) {
	var object *unstructured.Unstructured
	var err error
	if len(objects) == 0 {
		clientSet := GetDynamicClientSet()
		if clientSet == nil {
			return false, "", fmt.Errorf("Error in getting clientset before fetching RouteBackendExtension CR object")
		}
		object, err = clientSet.Resource(RouteBackendExtensionCRDGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, "", fmt.Errorf("RouteBackendExtension object %s/%s not found", namespace, name)
			}
			return false, "", err
		}
	} else {
		object = objects[0]
	}
	statusJSON, found, err := unstructured.NestedMap(object.UnstructuredContent(), "status")
	if err != nil || !found {
		utils.AviLog.Warnf("key: %s, msg: Status not found for RouteBackendExtension CR %s/%s, err: %+v", key, namespace, name, err)
		return false, "", fmt.Errorf("Status not found for RouteBackendExtension CR %s/%s", namespace, name)
	}
	// fetch the status
	status, ok := statusJSON["status"]
	if !ok || status == "" {
		utils.AviLog.Warnf("key: %s, msg: RouteBackendExtension CR %s/%s has an invalid status field", key, namespace, name)
		return false, "", fmt.Errorf("RouteBackendExtension CR %s/%s has an invalid status field", namespace, name)
	}
	controller, found, err := unstructured.NestedString(statusJSON, "controller")
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: RouteBackendExtension CR %s/%s controller status not found: %+v", key, namespace, name, err)
		return false, status.(string), fmt.Errorf("RouteBackendExtension CR %s/%s controller status not found", namespace, name)
	}
	if !found || controller == "" {
		return false, status.(string), fmt.Errorf("RouteBackendExtension CR %s/%s is not processed by AKO CRD Operator", namespace, name)
	}
	if controller != AKOCRDController {
		return false, status.(string), fmt.Errorf("RouteBackendExtension CR %s/%s is not handled by AKO CRD Operator", namespace, name)
	}
	return true, status.(string), nil
}

func IsL7CRDValid(key, namespace, name string) (bool, *unstructured.Unstructured, error) {
	clientSet := GetDynamicClientSet()
	if clientSet == nil {
		return false, nil, fmt.Errorf("error in fetching L7Rule CRD object. clientset is nil")
	}
	obj, err := clientSet.Resource(L7CRDGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			utils.AviLog.Errorf("key:%s/%s, msg:L7Rule CRD not found: %+v", namespace, name, err)
			return false, nil, fmt.Errorf("L7Rule CRD %s/%s not found", namespace, name)
		}
		return false, nil, err
	}
	statusJSON, found, err := unstructured.NestedMap(obj.UnstructuredContent(), "status")
	if err != nil || !found {
		utils.AviLog.Errorf("key:%s/%s, msg:L7Rule CRD status not found: %+v", namespace, name, err)
		return false, nil, err
	}
	status, ok := statusJSON["status"]
	if !ok || status.(string) == "" {
		utils.AviLog.Errorf("key:%s, msg: error: L7Rule CRD %s/%s is not processed by AKO main container", key, namespace, name)
		return false, nil, fmt.Errorf("L7Rule CRD %s/%s is not processed by AKO main container", namespace, name)
	}
	if status.(string) != lib.StatusAccepted {
		utils.AviLog.Errorf("key: %s, msg: error: L7Rule CRD %s/%s is not accepted", key, namespace, name)
		return false, nil, fmt.Errorf("L7Rule CRD %s/%s is not accepted", namespace, name)
	}
	return true, obj, nil
}

func ParseL7CRD(key, namespace, name string, vsNode nodes.AviVsEvhSniModel, isFilterAppProfSet bool) error {
	ok, obj, err := IsL7CRDValid(key, namespace, name)
	// second check is precautionary check to avoid failure in fetching content
	if !ok || obj == nil {
		return err
	}
	specJSON, found, err := unstructured.NestedMap(obj.UnstructuredContent(), "spec")
	if err != nil || !found {
		if err != nil {
			utils.AviLog.Warnf("key:%s/%s, msg:L7Rule CRD spec not found: %+v", namespace, name, err)
			return err
		}
		return fmt.Errorf("key: %s/%s, msg: L7Rule CRD spec not found", namespace, name)
	}
	generatedFields := vsNode.GetGeneratedFields()

	// bot policy has license restriction
	generatedFields.BotPolicyRef = getFieldValueTypeString(specJSON, "botPolicyRef")
	generatedFields.TrafficCloneProfileRef = getFieldValueTypeString(specJSON, "trafficCloneProfileRef")
	generatedFields.AllowInvalidClientCert = getFieldValueTypeBool(specJSON, "allowInvalidClientCert")
	generatedFields.CloseClientConnOnConfigUpdate = getFieldValueTypeBool(specJSON, "closeClientConnOnConfigUpdate")
	generatedFields.IgnPoolNetReach = getFieldValueTypeBool(specJSON, "ignPoolNetReach")
	generatedFields.RemoveListeningPortOnVsDown = getFieldValueTypeBool(specJSON, "removeListeningPortOnVsDown")
	generatedFields.MinPoolsUp = getFieldValueTypeUint32(specJSON, "minPoolsUp")
	generatedFields.SslSessCacheAvgSize = getFieldValueTypeUint32(specJSON, "sslSessCacheAvgSize")

	generatedFields.ConvertToRef()

	vsAnalyticsProfile := getFieldValueFromSpec(specJSON, "analyticsProfile")
	if vsAnalyticsProfile != nil {
		vsNode.SetAnalyticsProfileRef(vsAnalyticsProfile)
	}

	// WAF Policy
	vsWafPolicy := getFieldValueFromSpec(specJSON, "wafPolicy")
	if vsWafPolicy != nil {
		vsNode.SetWafPolicyRef(vsWafPolicy)
	}

	//ICAP profile
	icapProfileObj, found, err := unstructured.NestedStringMap(specJSON, "icapProfile")
	if err == nil && found {
		//get the object
		vsICAPProfile := []string{fmt.Sprintf("/api/icapprofile?name=%s", icapProfileObj["name"])}
		vsNode.SetICAPProfileRefs(vsICAPProfile)
	}
	//ErrorPage Profile
	errorProfileObj, found, err := unstructured.NestedStringMap(specJSON, "errorPageProfile")
	if err == nil && found {
		//get the object
		vsErrorPageProfile := fmt.Sprintf("/api/errorpageprofile?name=%s", errorProfileObj["name"])
		vsNode.SetErrorPageProfileRef(vsErrorPageProfile)
	}

	//Application Profile -- set only when there is no AppProf on filter
	if !isFilterAppProfSet {
		vsApplicationProfile := getFieldValueFromSpec(specJSON, "applicationProfile")
		if vsApplicationProfile != nil {
			vsNode.SetAppProfileRef(vsApplicationProfile)
		}
	}

	// Fetch HTTPPolicyset - can be directly fetched with call unstructured.NestedStringSlice(specJSON,"httpPolicy",  "policySets")
	httpPolicyObj, found, err := unstructured.NestedMap(specJSON, "httpPolicy")
	if err == nil && found {
		policysetObj, found, err := unstructured.NestedStringSlice(httpPolicyObj, "policySets")
		if err == nil && found {
			var validPolicySets []string
			// avoid duplicate and empty strings
			for _, v := range policysetObj {
				if trimmed := strings.TrimSpace(v); trimmed != "" {
					validPolicySets = append(validPolicySets, trimmed)
				}
			}
			httpPolicySet := sets.NewString(validPolicySets...).List()
			vsHTTPPolicySets := make([]string, len(httpPolicySet))
			for i, v := range httpPolicySet {
				vsHTTPPolicySets[i] = fmt.Sprintf("/api/httppolicyset?name=%s", v)
			}
			vsNode.SetHttpPolicySetRefs(vsHTTPPolicySets)
		}
		// overwrite AKO-GatewayAPI created HTTP Policysets
		overwrite, found, err := unstructured.NestedBool(httpPolicyObj, "overwrite")
		if err == nil && found {
			if overwrite {
				vsNode.SetHttpPolicyRefs([]*nodes.AviHttpPolicySetNode{})
			}
		}
	}

	// AnalyticsPolicy
	analyticPolicyObj, found, err := unstructured.NestedMap(specJSON, "analyticsPolicy")
	if err == nil && found {
		analyticsPolicy := &models.AnalyticsPolicy{}
		fullClientLogObj, found, err := unstructured.NestedMap(analyticPolicyObj, "fullClientLogs")
		if err == nil && found {
			enabled, found, err := unstructured.NestedBool(fullClientLogObj, "enabled")
			if err == nil && found && enabled {
				analyticsPolicy.FullClientLogs = &models.FullClientLogs{
					Enabled: &enabled,
				}
				duration, found, err := unstructured.NestedInt64(fullClientLogObj, "duration")
				if err == nil && found {
					analyticsPolicy.FullClientLogs.Duration = proto.Uint32(uint32(duration))
				}
				throttle, found, err := unstructured.NestedString(fullClientLogObj, "throttle")
				if err == nil && found {
					analyticsPolicy.FullClientLogs.Throttle = lib.GetThrottle(throttle)
				}
			}

		}
		allHeaders, found, err := unstructured.NestedBool(analyticPolicyObj, "logAllHeaders")
		if err == nil && found {
			analyticsPolicy.AllHeaders = &allHeaders
		}
		vsNode.SetAnalyticsPolicy(analyticsPolicy)
	}
	return nil
}
func getFieldValueTypeString(specJSON map[string]interface{}, fieldName string) *string {
	var value *string
	obj, found, err := unstructured.NestedString(specJSON, fieldName)
	if err == nil && found {
		return proto.String(obj)
	}
	return value
}

func getFieldValueTypeUint32(specJSON map[string]interface{}, fieldName string) *uint32 {
	var value *uint32
	obj, found, err := unstructured.NestedInt64(specJSON, fieldName)
	if err == nil && found {
		// safe side adding this check. schema validation should prevent this.
		if obj >= 0 && obj <= math.MaxUint32 {
			return proto.Uint32(uint32(obj))
		}
		utils.AviLog.Warnf("Field %s value %d is out of bounds for uint32", fieldName, obj)
	}
	return value
}
func getFieldValueTypeBool(specJSON map[string]interface{}, fieldName string) *bool {
	var value *bool
	obj, found, err := unstructured.NestedBool(specJSON, fieldName)
	if err == nil && found {
		return proto.Bool(obj)
	}
	return value
}
func getFieldValueFromSpec(specJSON map[string]interface{}, fieldName string) *string {
	var value *string
	obj, found, err := unstructured.NestedStringMap(specJSON, fieldName)
	if err == nil && found {
		vsField := proto.String(fmt.Sprintf("/api/%s?name=%s", strings.ToLower(fieldName), obj["name"]))
		return vsField
	}
	return value
}

func ParseRouteBackendExtensionCR(key, namespace, name string, poolNode *nodes.AviPoolNode, isFilterHMSet bool) error {
	clientSet := GetDynamicClientSet()
	if clientSet == nil {
		return fmt.Errorf("Error in getting clientset before fetching RouteBackendExtension CR object")
	}
	obj, err := clientSet.Resource(RouteBackendExtensionCRDGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("RouteBackendExtension CR %s/%s not found", namespace, name)
		}
		return err
	}
	_, status, err := IsRouteBackendExtensionProcessed(key, namespace, name, obj)
	if status != lib.StatusAccepted || err != nil {
		return fmt.Errorf("RouteBackendExtension CR %s/%s is not accepted", namespace, name)
	}
	specJSON, found, err := unstructured.NestedMap(obj.UnstructuredContent(), "spec")
	if err != nil || !found {
		return fmt.Errorf("RouteBackendExtension CR %s/%s spec not found", namespace, name)
	}

	if lbAlgo, found, err := unstructured.NestedString(specJSON, "lbAlgorithm"); err == nil && found {
		poolNode.LbAlgorithm = &lbAlgo
	}
	if lbAlgoHash, found, err := unstructured.NestedString(specJSON, "lbAlgorithmHash"); err == nil && found {
		poolNode.LbAlgorithmHash = &lbAlgoHash
	}
	if lbAlgoHashHdr, found, err := unstructured.NestedString(specJSON, "lbAlgorithmConsistentHashHdr"); err == nil && found {
		poolNode.LbAlgorithmConsistentHashHdr = &lbAlgoHashHdr
	}
	// if sessionPersistence is not set in rule, set it from the spec of routebackendextension
	if poolNode.ApplicationPersistenceProfile == nil {
		if persistenceProfile, found, err := unstructured.NestedString(specJSON, "persistenceProfile"); err == nil && found {
			poolNode.ApplicationPersistenceProfileRef = proto.String(fmt.Sprintf("/api/applicationpersistenceprofile?name=%s", persistenceProfile))
		}
	}
	if !isFilterHMSet {
		hms, found, err := unstructured.NestedSlice(specJSON, "healthMonitor")
		if err == nil && found {
			for _, hm := range hms {
				if hmMap, ok := hm.(map[string]interface{}); ok {
					hmName, found, err := unstructured.NestedString(hmMap, "name")
					if err == nil && found {
						hmRef := proto.String(fmt.Sprintf("/api/healthmonitor?name=%s", hmName))
						poolNode.HealthMonitorRefs = append(poolNode.HealthMonitorRefs, *hmRef)
					}
				}
			}
		}
	}

	// Parse SSL/TLS related fields from backendTLS
	if backendTLSMap, found, err := unstructured.NestedMap(specJSON, "backendTLS"); err == nil && found {

		poolNode.SslProfileRef = proto.String(fmt.Sprintf("/api/sslprofile?name=%s", lib.DefaultPoolSSLProfile))

		// Parse hostCheckEnabled
		if hostCheckEnabled, found, err := unstructured.NestedBool(backendTLSMap, "hostCheckEnabled"); err == nil && found {
			poolNode.HostCheckEnabled = proto.Bool(hostCheckEnabled)
		}

		// Parse domainName
		if domainNames, found, err := unstructured.NestedStringSlice(backendTLSMap, "domainName"); err == nil && found && len(domainNames) > 0 {
			// Use all domain names from the spec
			poolNode.DomainName = domainNames
		}

		// Parse PKIProfile
		if pkiProfileMap, found, err := unstructured.NestedMap(backendTLSMap, "pkiProfile"); err == nil && found {
			if pkiProfileKind, found, err := unstructured.NestedString(pkiProfileMap, "kind"); err == nil && found && pkiProfileKind == "CRD" {
				if pkiProfileName, found, err := unstructured.NestedString(pkiProfileMap, "name"); err == nil && found {
					// Set PKI profile reference using clustername-namespace-name format - this will be used by the pool node for backend SSL validation
					pkiProfileFullName := getPKIProfileName(namespace, pkiProfileName)
					poolNode.PkiProfileRef = proto.String(fmt.Sprintf("/api/pkiprofile?name=%s", pkiProfileFullName))
				}
			}
		}
	}

	return nil

}

// IsApplicationProfileValid checks if the ApplicationProfile CRD is valid as well as ready and processed by AKO CRD Operator.
func IsApplicationProfileValid(namespace, name string) (bool, bool) {
	clientSet := GetDynamicClientSet()
	obj, err := clientSet.Resource(AppProfileCRDGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD status not found: %+v", namespace, name, err)
		}
		return false, false
	}

	return true, IsApplicationProfileProcessed(obj, namespace, name)
}

// IsApplicationProfileProcessed checks if the ApplicationProfile CRD is processed by AKO CRD operator.
// It returns true if the ApplicationProfile's backendname is updated and status is "Programmed", false otherwise.
// It also returns the status of the ApplicationProfile.
func IsApplicationProfileProcessed(obj interface{}, namespace, name string) bool {
	oldObj := obj.(*unstructured.Unstructured)
	statusJSON, found, err := unstructured.NestedMap(oldObj.UnstructuredContent(), "status")
	if err != nil || !found {
		utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD status not found: %+v", namespace, name, err)
		return false
	}

	tenant, ok := statusJSON["tenant"]
	if !ok || tenant == "" {
		utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD tenant not found", namespace, name)
		return false
	}
	namespaceTenant := lib.GetTenantInNamespace(namespace)
	if tenant != namespaceTenant {
		utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD tenant %s is not same as namespace tenant %s", namespace, name, tenant, namespaceTenant)
		return false
	}

	// fetch the backendObjectName
	backendObjectName, ok := statusJSON["backendObjectName"]
	if !ok || backendObjectName == "" {
		utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD backendObjectName not found", namespace, name)
		return false
	}

	conditions, ok := statusJSON["conditions"].([]interface{})
	if !ok || len(conditions) == 0 {
		utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD conditions not found", namespace, name)
		return false
	}

	for _, c := range conditions {
		condition := c.(map[string]interface{})
		statusReady := condition["status"].(string) == string(metav1.ConditionTrue)
		conditionReady := condition["type"].(string) == "Programmed"

		if statusReady && conditionReady {
			return true
		}
	}

	utils.AviLog.Warnf("key:%s/%s, msg: ApplicationProfile CRD is not ready", namespace, name)
	return false
}

// getPKIProfileName generates PKI profile name using clustername-namespace-name format
// to match the naming convention used by the ako-crd-operator
func getPKIProfileName(namespace, objectName string) string {
	name := namespace + "-" + objectName
	namePrefix := CRDOperatorPrefix + lib.GetClusterName() + "--"
	return lib.EncodeWithPrefix(name, lib.EVHVS, namePrefix)
}

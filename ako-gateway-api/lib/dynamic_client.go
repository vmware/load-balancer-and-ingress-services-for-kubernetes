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
	L7CRDInformer informers.GenericInformer
}

// NewDynamicInformers initializes the DynamicInformers struct
func NewDynamicInformers(client dynamic.Interface, akoInfra bool) *DynamicInformers {
	informers := &DynamicInformers{}
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)

	// not applicable in wcp context
	if !utils.IsWCP() {
		informers.L7CRDInformer = f.ForResource(L7CRDGVR)
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

func ParseL7CRD(key, namespace, name string, vsNode nodes.AviVsEvhSniModel, isFilterAppProfSet bool) error {

	clientSet := GetDynamicClientSet()
	if clientSet == nil {
		return fmt.Errorf("key: %s, msg:error in fetching L7Rule CRD object", key)
	}
	obj, err := clientSet.Resource(L7CRDGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("key: %s, msg: error: L7Rule CRD %s/%s not found", key, namespace, name)
		}
		return err
	}
	statusJSON, found, err := unstructured.NestedMap(obj.UnstructuredContent(), "status")
	if err != nil || !found {
		utils.AviLog.Warnf("key:%s/%s, msg:L7Rule CRD status not found: %+v", namespace, name, err)
		return err
	}
	status, ok := statusJSON["status"]
	if !ok || status.(string) == "" {
		return fmt.Errorf("key:%s, msg: error: L7Rule CRD %s/%s is not processed by AKO main container", key, namespace, name)
	}
	if status.(string) != lib.StatusAccepted {
		return fmt.Errorf("key: %s, msg: error: L7Rule CRD %s/%s is not accepted", key, namespace, name)
	}
	specJSON, found, err := unstructured.NestedMap(obj.UnstructuredContent(), "spec")
	if err != nil || !found {
		utils.AviLog.Warnf("key:%s/%s, msg:L7Rule CRD spec not found: %+v", namespace, name, err)
		return err
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
			// avoid duplicates
			httpPolicySet := sets.NewString(policysetObj...).List()
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
		return proto.Uint32(uint32(obj))
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

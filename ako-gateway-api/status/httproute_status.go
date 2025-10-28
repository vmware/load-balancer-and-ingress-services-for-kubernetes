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

package status

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type httproute struct{}

func (o *httproute) Get(key string, name string, namespace string) *gatewayv1.HTTPRoute {

	obj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the HTTPRoute object. err: %s", key, err)
		return nil
	}
	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the HTTPRoute object %s", key, name)
	return obj.DeepCopy()
}

func (o *httproute) GetAll(key string) map[string]*gatewayv1.HTTPRoute {

	objs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the HTTPRoute objects. err: %s", key, err)
		return nil
	}

	httpRouteMap := make(map[string]*gatewayv1.HTTPRoute)
	for _, obj := range objs {
		httpRouteMap[obj.Namespace+"/"+obj.Name] = obj.DeepCopy()
	}

	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the HTTPRoute objects", key)
	return httpRouteMap
}

func (o *httproute) Delete(key string, option status.StatusOptions) {
	nsName := strings.Split(option.Options.ServiceMetadata.HTTPRoute, "/")
	if len(nsName) != 2 {
		utils.AviLog.Warnf("key: %s, msg: invalid HttpRoute name and namespace", key)
		return
	}
	namespace := nsName[0]
	name := nsName[1]
	httpRoute := o.Get(key, name, namespace)
	if httpRoute != nil {
		// Update HTTPRoute status to remove VS UUID
		if err := o.removeVSUUIDFromHTTPRouteStatus(key, httpRoute, option.Options); err != nil {
			utils.AviLog.Warnf("key: %s, msg: failed to remove VS UUID from HTTPRoute status: %v", key, err)
		}
	}
}

func (o *httproute) Update(key string, option status.StatusOptions) {
	nsName := strings.Split(option.Options.ServiceMetadata.HTTPRoute, "/")
	if len(nsName) != 2 {
		utils.AviLog.Warnf("key: %s, msg: invalid HttpRoute name and namespace", key)
		return
	}
	namespace := nsName[0]
	name := nsName[1]
	httpRoute := o.Get(key, name, namespace)
	if httpRoute != nil {
		if option.Options.Status != nil {
			option.Options.Status.HTTPRouteStatus = akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(lib.HTTPRoute + "/" + namespace + "/" + name)
			o.Patch(key, httpRoute, option.Options.Status)
		}
		// Update HTTPRoute status Accepted condition with VS UUID
		if err := o.updateHTTPRouteStatusWithVSUUID(key, httpRoute, option.Options); err != nil {
			utils.AviLog.Warnf("key: %s, msg: failed to update HTTPRoute status with VS UUID: %v", key, err)
		}
	}
}

func (o *httproute) BulkUpdate(key string, options []status.StatusOptions) {
	httpRouteMap := o.GetAll(key)
	for _, option := range options {
		httpRoute := httpRouteMap[option.Options.ServiceMetadata.HTTPRoute]
		if httpRoute != nil {
			if err := o.updateHTTPRouteStatusWithVSUUID(key, httpRoute, option.Options); err != nil {
				utils.AviLog.Warnf("key: %s, msg: failed to update HTTPRoute status with VS UUID: %v", key, err)
			}
		}
	}
}

func (o *httproute) Patch(key string, obj runtime.Object, status *status.Status, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 5 {
			utils.AviLog.Errorf("key: %s, msg: Patch retried 5 times, aborting", key)
			akogatewayapilib.AKOControlConfig().EventRecorder().Eventf(obj, corev1.EventTypeWarning, lib.PatchFailed, "Patch of status failed after multiple retries")
			return errors.New("Patch retried 5 times, aborting")
		}
	}

	httpRoute := obj.(*gatewayv1.HTTPRoute)
	if o.isStatusEqual(&httpRoute.Status, status.HTTPRouteStatus) {
		return nil
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": status.HTTPRouteStatus,
	})
	_, err := akogatewayapilib.AKOControlConfig().GatewayAPIClientset().GatewayV1().HTTPRoutes(httpRoute.Namespace).Patch(context.TODO(), httpRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was an error in updating the HTTPRoute status. err: %+v, retry: %d", key, err, retry)
		updatedObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(httpRoute.Namespace).Get(httpRoute.Name)
		if err != nil {
			utils.AviLog.Warnf("HTTPRoute not found %v", err)
			return err
		}
		return o.Patch(key, updatedObj, status, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the HTTPRoute %s/%s status %+v", key, httpRoute.Namespace, httpRoute.Name, utils.Stringify(status))
	return nil
}

func (o *httproute) isStatusEqual(old, new *gatewayv1.HTTPRouteStatus) bool {
	oldStatus, newStatus := old.DeepCopy(), new.DeepCopy()
	currentTime := metav1.Now()
	for i := range oldStatus.Parents {
		for j := range oldStatus.Parents[i].Conditions {
			oldStatus.Parents[i].Conditions[j].LastTransitionTime = currentTime
		}
	}
	for i := range newStatus.Parents {
		for j := range newStatus.Parents[i].Conditions {
			newStatus.Parents[i].Conditions[j].LastTransitionTime = currentTime
		}
	}
	return reflect.DeepEqual(oldStatus, newStatus)
}

// updateHTTPRouteStatusWithVSUUID updates HTTPRoute status with VS UUID
func (o *httproute) updateHTTPRouteStatusWithVSUUID(key string, httpRoute *gatewayv1.HTTPRoute, options *status.UpdateOptions) error {
	// Loop over route parent status and match with gateway name
	gatewayNSName := options.ServiceMetadata.Gateway
	ruleName := options.ServiceMetadata.HTTPRouteRuleName
	virtualServiceUUID := options.VirtualServiceUUID

	if gatewayNSName == "" || ruleName == "" || virtualServiceUUID == "" {
		utils.AviLog.Debugf("key: %s, msg: Missing required fields for HTTPRoute status update - Gateway: %s, RuleName: %s, VSUUID: %s", key, gatewayNSName, ruleName, virtualServiceUUID)
		return nil
	}

	// Parse gateway namespace and name
	gatewayParts := strings.Split(gatewayNSName, "/")
	if len(gatewayParts) != 2 {
		return fmt.Errorf("invalid gateway name format: %s", gatewayNSName)
	}
	gatewayNamespace := gatewayParts[0]
	gatewayName := gatewayParts[1]

	// Create or update HTTPRoute status
	httpRouteStatus := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(lib.HTTPRoute + "/" + options.ServiceMetadata.HTTPRoute)
	if httpRouteStatus.Parents == nil {
		httpRouteStatus.Parents = []gatewayv1.RouteParentStatus{}
	}

	// Find or create parent status for this gateway
	for i := range httpRouteStatus.Parents {
		if string(httpRouteStatus.Parents[i].ParentRef.Name) == gatewayName &&
			(httpRouteStatus.Parents[i].ParentRef.Namespace == nil || string(*httpRouteStatus.Parents[i].ParentRef.Namespace) == gatewayNamespace) {
			parentStatus := &httpRouteStatus.Parents[i]

			// Add VSUUID only if the HTTPRoute is Accepted for this Gateway
			for _, condition := range parentStatus.Conditions {
				if condition.Type == string(gatewayv1.RouteConditionAccepted) && condition.Status == metav1.ConditionTrue {
					message, err := o.buildJSONMessage(parentStatus.Conditions, ruleName, virtualServiceUUID, false)
					if err != nil {
						return err
					}
					newCondition := NewCondition().
						Type(string(gatewayv1.RouteConditionAccepted)).
						Status(metav1.ConditionTrue).
						Reason(string(gatewayv1.RouteReasonAccepted)).
						ObservedGeneration(httpRoute.ObjectMeta.Generation).
						Message(message)
					newCondition.SetIn(&parentStatus.Conditions)
				}
			}
		}
	}

	akogatewayapiobjects.GatewayApiLister().UpdateRouteToRouteStatusMapping(lib.HTTPRoute+"/"+options.ServiceMetadata.HTTPRoute, httpRouteStatus)
	// Patch the HTTPRoute status
	return o.Patch(key, httpRoute, &status.Status{HTTPRouteStatus: httpRouteStatus})
}

// removeVSUUIDFromHTTPRouteStatus updates HTTPRoute status to remove VS UUID
func (o *httproute) removeVSUUIDFromHTTPRouteStatus(key string, httpRoute *gatewayv1.HTTPRoute, options *status.UpdateOptions) error {
	// Loop over route parent status and match with gateway name
	gatewayNSName := options.ServiceMetadata.Gateway
	ruleName := options.ServiceMetadata.HTTPRouteRuleName

	if gatewayNSName == "" {
		utils.AviLog.Debugf("key: %s, msg: Missing gateway name for HTTPRoute status delete", key)
		return nil
	}

	// Parse gateway namespace and name
	gatewayParts := strings.Split(gatewayNSName, "/")
	if len(gatewayParts) != 2 {
		return fmt.Errorf("invalid gateway name format: %s", gatewayNSName)
	}
	gatewayNamespace := gatewayParts[0]
	gatewayName := gatewayParts[1]

	// Update HTTPRoute status
	httpRouteStatus := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(lib.HTTPRoute + "/" + options.ServiceMetadata.HTTPRoute)
	if httpRouteStatus.Parents == nil {
		utils.AviLog.Debugf("key: %s, msg: No parents to update for HTTPRoute status delete", key)
		return nil // No parents to update
	}

	// Find parent status for this gateway
	for i := range httpRouteStatus.Parents {
		if string(httpRouteStatus.Parents[i].ParentRef.Name) == gatewayName &&
			(httpRouteStatus.Parents[i].ParentRef.Namespace == nil || string(*httpRouteStatus.Parents[i].ParentRef.Namespace) == gatewayNamespace) {

			// Add VSUUID only if the HTTPRoute is Accepted for this Gateway
			for _, condition := range httpRouteStatus.Parents[i].Conditions {
				if condition.Type == string(gatewayv1.RouteConditionAccepted) && condition.Status == metav1.ConditionTrue {
					message, err := o.buildJSONMessage(httpRouteStatus.Parents[i].Conditions, ruleName, "", true)
					if err != nil {
						return err
					}
					status := metav1.ConditionTrue
					reason := string(gatewayv1.RouteReasonAccepted)
					if message == "" {
						status = metav1.ConditionFalse
						reason = string(gatewayv1.RouteReasonPending)
					}
					newCondition := NewCondition().
						Type(string(gatewayv1.RouteConditionAccepted)).
						Status(status).
						Reason(reason).
						ObservedGeneration(httpRoute.ObjectMeta.Generation).
						Message(message)
					newCondition.SetIn(&httpRouteStatus.Parents[i].Conditions)
				}
			}
		}
	}

	akogatewayapiobjects.GatewayApiLister().UpdateRouteToRouteStatusMapping(lib.HTTPRoute+"/"+options.ServiceMetadata.HTTPRoute, httpRouteStatus)
	// Patch the HTTPRoute status
	return o.Patch(key, httpRoute, &status.Status{HTTPRouteStatus: httpRouteStatus})
}

// buildJSONMessage builds or updates a JSON message for HTTPRoute conditions
// If isDelete is true, it removes the ruleName from the message
// If isDelete is false, it adds/updates the ruleName with virtualServiceUUID
func (o *httproute) buildJSONMessage(conditions []metav1.Condition, ruleName, virtualServiceUUID string, isDelete bool) (string, error) {
	// Find existing Accepted condition to get current message
	existingMessage := ""
	for _, condition := range conditions {
		if condition.Type == string(gatewayv1.RouteConditionAccepted) {
			existingMessage = condition.Message
			break
		}
	}

	// Parse existing message as JSON map
	ruleVSMap := make(map[string]string)
	if existingMessage != "" && existingMessage != "Parent reference is valid" {
		// Try to unmarshal as JSON first
		if err := json.Unmarshal([]byte(existingMessage), &ruleVSMap); err != nil {
			return "", err
		}
	}

	if isDelete {
		// Remove the rule from the map
		delete(ruleVSMap, ruleName)

		// If map is empty, return deletion message
		if len(ruleVSMap) == 0 {
			return "", nil
		}
	} else {
		// Add or update the rule in the map
		ruleVSMap[ruleName] = virtualServiceUUID
	}

	// Marshal the map to JSON
	messageBytes, err := json.Marshal(ruleVSMap)
	if err != nil {
		return "", err
	}

	return string(messageBytes), nil
}

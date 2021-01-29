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

package status

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	svcapiv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type UpdateSvcApiGWStatusConditionOptions struct {
	Type    string                 // to be casted to the appropriate conditionType
	Status  metav1.ConditionStatus // True/False/Unknown
	Message string                 // extended condition message
	Reason  string                 // reason for transition
}

func UpdateSvcApiGatewayStatusAddress(options []UpdateStatusOptions, bulk bool) {
	gatewaysToUpdate, updateGWOptions := parseOptionsFromMetadata(options, bulk)
	var updateServiceOptions []UpdateStatusOptions

	// gatewayMap: {ns/gateway: gatewayObj}
	// this pre-fetches all gateways to be candidates for status update
	// after pre-fetching, if a status update comes for that gateway, then the pre-fetched gateway would be stale
	// in which case gateway will be fetched again in updateObject, as part of a retry
	gatewayMap := getSvcApiGateways(gatewaysToUpdate, bulk)
	for _, option := range updateGWOptions {
		updateServiceOptions = append(updateServiceOptions, UpdateStatusOptions{
			Vip: option.Vip,
			Key: option.Key,
			ServiceMetadata: avicache.ServiceMetadataObj{
				NamespaceServiceName: option.ServiceMetadata.NamespaceServiceName,
			},
		})

		if gw := gatewayMap[option.IngSvc]; gw != nil {
			// assuming 1 IP per gateway
			gwStatus := gw.Status.DeepCopy()
			gwStatus.Addresses = []svcapiv1alpha1.GatewayAddress{{
				Value: option.Vip,
				Type:  svcapiv1alpha1.IPAddressType,
			}}

			// when statuses are synced during bootup
			InitializeSvcApiGatewayConditions(gwStatus, &gw.Spec, true)
			UpdateSvcApiGatewayStatusGWCondition(option.Key, gwStatus, &UpdateSvcApiGWStatusConditionOptions{
				Type:   "Ready",
				Status: metav1.ConditionTrue,
			})
			UpdateSvcApiGatewayStatusObject(option.Key, gw, gwStatus)
		}
	}

	UpdateL4LBStatus(updateServiceOptions, bulk)
	return
}

// getGateways fetches all ingresses and returns a map: {"namespace/name": ingressObj...}
// if bulk is set to true, this fetches all ingresses in a single k8s api-server call
func getSvcApiGateways(gwNSNames []string, bulk bool, retryNum ...int) map[string]*svcapiv1alpha1.Gateway {
	retry := 0
	gwMap := make(map[string]*svcapiv1alpha1.Gateway)
	if len(retryNum) > 0 {
		utils.AviLog.Infof("Retrying to get the gateway for status update")
		retry = retryNum[0]
		if retry >= 2 {
			utils.AviLog.Errorf("getGateways for status update retried 3 times, aborting")
			return gwMap
		}
	}

	if bulk {
		gwList, err := lib.GetServicesAPIClientset().NetworkingV1alpha1().Gateways("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the gateway object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getSvcApiGateways(gwNSNames, bulk, retry+1)
			}
		} else {
			for i := range gwList.Items {
				ing := gwList.Items[i]
				gwMap[ing.Namespace+"/"+ing.Name] = &ing
			}
		}
		return gwMap
	}

	for _, namespaceName := range gwNSNames {
		nsNameSplit := strings.Split(namespaceName, "/")
		gw, err := lib.GetServicesAPIClientset().NetworkingV1alpha1().Gateways(nsNameSplit[0]).Get(context.TODO(), nsNameSplit[1], metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the gateway object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getSvcApiGateways(gwNSNames, bulk, retry+1)
			}
		} else {
			gwMap[gw.Namespace+"/"+gw.Name] = gw
		}

	}

	return gwMap
}

func DeleteSvcApiGatewayStatusAddress(key string, svcMetadataObj avicache.ServiceMetadataObj) error {
	gwNSName := strings.Split(svcMetadataObj.Gateway, "/")
	gw, err := lib.GetServicesAPIClientset().NetworkingV1alpha1().Gateways(gwNSName[0]).Get(context.TODO(), gwNSName[1], metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was a problem in resetting the gateway address status: %s", key, err)
		return err
	}

	if len(gw.Status.Addresses) == 0 ||
		(len(gw.Status.Addresses) > 0 && gw.Status.Addresses[0].Value == "") {
		return nil
	}

	// assuming 1 IP per gateway
	gwStatus := gw.Status.DeepCopy()
	gwStatus.Addresses = []svcapiv1alpha1.GatewayAddress{}
	UpdateSvcApiGatewayStatusGWCondition(key, gwStatus, &UpdateSvcApiGWStatusConditionOptions{
		Type:   "Pending",
		Status: metav1.ConditionTrue,
		Reason: "virtualservice deleted/notfound",
	})
	UpdateSvcApiGatewayStatusObject(key, gw, gwStatus)

	utils.AviLog.Infof("key: %s, msg: Successfully reset the address status of gateway: %s", key, svcMetadataObj.Gateway)
	return nil
}

func DeleteSvcApiStatus(key string, svcMetadataObj avicache.ServiceMetadataObj) error {
	gwNSName := strings.Split(svcMetadataObj.Gateway, "/")
	gw, err := lib.GetServicesAPIClientset().NetworkingV1alpha1().Gateways(gwNSName[0]).Get(context.TODO(), gwNSName[1], metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was a problem in resetting the gateway address status: %s", key, err)
		return err
	}

	if len(gw.Status.Addresses) == 0 ||
		(len(gw.Status.Addresses) > 0 && gw.Status.Addresses[0].Value == "") {
		return nil
	}

	// assuming 1 IP per gateway
	gwStatus := gw.Status.DeepCopy()
	gwStatus.Addresses = []svcapiv1alpha1.GatewayAddress{}
	UpdateSvcApiGatewayStatusGWCondition(key, gwStatus, &UpdateSvcApiGWStatusConditionOptions{
		Type:   "Pending",
		Status: metav1.ConditionTrue,
		Reason: "virtualservice deleted/notfound",
	})
	UpdateSvcApiGatewayStatusObject(key, gw, gwStatus)
	return nil
}

// supported GatewayConditionTypes
// InvalidListeners, InvalidAddress, *Serviceable
func UpdateSvcApiGatewayStatusGWCondition(key string, gwStatus *svcapiv1alpha1.GatewayStatus, updateStatus *UpdateSvcApiGWStatusConditionOptions) {
	utils.AviLog.Debugf("key: %s, msg: Updating Gateway status gateway condition %v", key, utils.Stringify(updateStatus))
	InitializeSvcApiGatewayConditions(gwStatus, nil, false)

	for i := range gwStatus.Conditions {
		if string(gwStatus.Conditions[i].Type) == updateStatus.Type {
			gwStatus.Conditions[i].Status = updateStatus.Status
			gwStatus.Conditions[i].Message = updateStatus.Message
			gwStatus.Conditions[i].Reason = updateStatus.Reason
			gwStatus.Conditions[i].LastTransitionTime = metav1.Now()
		}

		var inverseCondition metav1.ConditionStatus
		if updateStatus.Status == metav1.ConditionFalse {
			inverseCondition = metav1.ConditionTrue
		} else {
			inverseCondition = metav1.ConditionFalse
		}
		// if Pending true, mark Ready as false automatically...
		// if Ready true, mark Pending as false automatically...
		if (updateStatus.Type == "Pending" && string(gwStatus.Conditions[i].Type) == "Ready") ||
			(updateStatus.Type == "Ready" && string(gwStatus.Conditions[i].Type) == "Pending") {
			gwStatus.Conditions[i].Status = inverseCondition
			gwStatus.Conditions[i].LastTransitionTime = metav1.Now()
			gwStatus.Conditions[i].Message = ""
			gwStatus.Conditions[i].Reason = ""
		}
	}

	var listenerConditionStatus metav1.ConditionStatus
	if updateStatus.Type == "Ready" {
		listenerConditionStatus = metav1.ConditionTrue
	} else {
		listenerConditionStatus = metav1.ConditionFalse
	}
	UpdateSvcApiGatewayStatusListenerConditions(key, gwStatus, "", &UpdateSvcApiGWStatusConditionOptions{
		Type:   "Ready",
		Status: listenerConditionStatus,
		Reason: "Ready",
	})
}

// supported ListenerConditionType
// PortConflict, InvalidRoutes, UnsupportedProtocol, *Serviceable
// pass portString as empty string for updating status in all ports
func UpdateSvcApiGatewayStatusListenerConditions(key string, gwStatus *svcapiv1alpha1.GatewayStatus, portString string, updateStatus *UpdateSvcApiGWStatusConditionOptions) {
	utils.AviLog.Debugf("key: %s, msg: Updating Gateway status listener condition port: %s %v", key, portString, utils.Stringify(updateStatus))
	for port, condition := range gwStatus.Listeners {
		notFound := true
		if portString == "" {
			for i, portCondition := range condition.Conditions {
				if updateStatus.Type == "Ready" && updateStatus.Type != string(portCondition.Type) && updateStatus.Status == metav1.ConditionTrue {
					gwStatus.Listeners[port].Conditions[i].Status = metav1.ConditionFalse
					gwStatus.Listeners[port].Conditions[i].Message = ""
					gwStatus.Listeners[port].Conditions[i].Reason = ""
				}

				if string(portCondition.Type) == updateStatus.Type {
					gwStatus.Listeners[port].Conditions[i].Status = updateStatus.Status
					gwStatus.Listeners[port].Conditions[i].Message = updateStatus.Message
					gwStatus.Listeners[port].Conditions[i].Reason = updateStatus.Reason
					gwStatus.Listeners[port].Conditions[i].LastTransitionTime = metav1.Now()
					notFound = false
				}
			}

			if notFound {
				gwStatus.Listeners[port].Conditions = append(gwStatus.Listeners[port].Conditions, metav1.Condition{
					Type:               updateStatus.Type,
					Status:             updateStatus.Status,
					Reason:             updateStatus.Reason,
					LastTransitionTime: metav1.Now(),
				})
			}
		}
	}

	// in case of a positive error listenerCondition Update we need to mark the
	// gateway Condition back from Ready to Pending
	badTypes := []string{"PortConflict", "InvalidRoutes", "UnsupportedProtocol"}
	if utils.HasElem(badTypes, updateStatus.Type) {
		UpdateSvcApiGatewayStatusGWCondition(key, gwStatus, &UpdateSvcApiGWStatusConditionOptions{
			Type:   "Pending",
			Status: metav1.ConditionTrue,
			Reason: fmt.Sprintf("port %s error %s", portString, updateStatus.Type),
		})
	}
}

func UpdateSvcApiGatewayStatusObject(key string, gw *svcapiv1alpha1.Gateway, updateStatus *svcapiv1alpha1.GatewayStatus, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 5 {
			utils.AviLog.Errorf("key: %s, msg: UpdateSvcApiGatewayStatusObject retried 5 times, aborting", key)
			return
		}
	}

	// if an IP address is present on the gateway object, it is fair to assume that the gateway corresponds to a VS in avi
	// in case the IP is not present, the gateway can be deleted freely since deleting that would be a NOOP for AKO
	// so we add finalizer when an IP is updated to the gateway, and remove it when we delete the IP address.
	if len(updateStatus.Addresses) > 0 && updateStatus.Addresses[0].Value != "" {
		lib.CheckAndSetSvcApiGatewayFinalizer(gw)
	} else {
		lib.RemoveSvcApiGatewayFinalizer(gw)
	}

	if compareSvcApiGatewayStatuses(&gw.Status, updateStatus) {
		return
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": updateStatus,
	})
	_, err := lib.GetServicesAPIClientset().NetworkingV1alpha1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Warnf("msg: %d there was an error in updating the gateway status: %+v", retry, err)
		updatedGW, err := lib.GetServicesAPIClientset().NetworkingV1alpha1().Gateways(gw.Namespace).Get(context.TODO(), gw.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("gateway not found %v", err)
			return
		}
		UpdateSvcApiGatewayStatusObject(key, updatedGW, updateStatus, retry+1)
	}

	utils.AviLog.Infof("msg: Successfully updated the gateway %s/%s status %+v", gw.Namespace, gw.Name, utils.Stringify(updateStatus))
	return
}

func InitializeSvcApiGatewayConditions(gwStatus *svcapiv1alpha1.GatewayStatus, gwSpec *svcapiv1alpha1.GatewaySpec, gwReady bool) {
	if len(gwStatus.Conditions) == 0 {
		gwStatus.Conditions = []metav1.Condition{{
			Type:               "Pending",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
		}, {
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
		}}
	}

	if gwSpec == nil {
		return
	}

	gwPortMap := make(map[svcapiv1alpha1.PortNumber][]metav1.Condition)
	for _, listenerStatus := range gwStatus.Listeners {
		gwPortMap[listenerStatus.Port] = listenerStatus.Conditions
	}

	var listenerStatuses []svcapiv1alpha1.ListenerStatus
	for _, listener := range gwSpec.Listeners {
		if val, ok := gwPortMap[svcapiv1alpha1.PortNumber(listener.Port)]; ok {
			listenerStatuses = append(listenerStatuses, svcapiv1alpha1.ListenerStatus{
				Port:       listener.Port,
				Conditions: val,
			})
		} else {
			var portCondition metav1.ConditionStatus
			if gwReady {
				portCondition = metav1.ConditionTrue
			} else {
				portCondition = metav1.ConditionFalse
			}
			listenerStatuses = append(listenerStatuses, svcapiv1alpha1.ListenerStatus{
				Port: listener.Port,
				Conditions: []metav1.Condition{{
					Type:               "Ready",
					Status:             portCondition,
					LastTransitionTime: metav1.Now(),
					Message:            "Initializing",
					Reason:             string(svcapiv1alpha1.GatewayReasonNotReconciled),
				}},
			})
		}
	}

	gwStatus.Listeners = listenerStatuses
	if len(gwStatus.Addresses) == 0 {
		gwStatus.Addresses = []svcapiv1alpha1.GatewayAddress{}
	}
	return
}

// do not compare lastTransitionTime updates in gateway
func compareSvcApiGatewayStatuses(old, new *svcapiv1alpha1.GatewayStatus) bool {
	oldStatus, newStatus := old.DeepCopy(), new.DeepCopy()
	currentTime := metav1.Now()
	for i := range oldStatus.Conditions {
		oldStatus.Conditions[i].LastTransitionTime = currentTime
	}
	for _, listener := range oldStatus.Listeners {
		for i := range listener.Conditions {
			listener.Conditions[i].LastTransitionTime = currentTime
		}
	}
	for i := range newStatus.Conditions {
		newStatus.Conditions[i].LastTransitionTime = currentTime
	}
	for _, listener := range newStatus.Listeners {
		for i := range listener.Conditions {
			listener.Conditions[i].LastTransitionTime = currentTime
		}
	}

	return reflect.DeepEqual(oldStatus, newStatus)
}

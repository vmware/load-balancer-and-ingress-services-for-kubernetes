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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

type UpdateGWStatusConditionOptions struct {
	Type    string               // to be casted to the appropriate conditionType
	Status  core.ConditionStatus // True/False/Unknown
	Message string               // extended condition message
	Reason  string               // reason for transition
}

func (l *leader) UpdateGatewayStatusAddress(options []UpdateOptions, bulk bool) {
	gatewaysToUpdate, updateGWOptions := parseOptionsFromMetadata(options, bulk)

	// gatewayMap: {ns/gateway: gatewayObj}
	// this pre-fetches all gateways to be candidates for status update
	// after pre-fetching, if a status update comes for that gateway, then the pre-fetched gateway would be stale
	// in which case gateway will be fetched again in updateObject, as part of a retry
	gatewayMap := getGateways(gatewaysToUpdate, bulk)
	skipDelete := map[string]bool{}
	for _, option := range updateGWOptions {
		if option.Message != "" {
			continue
		}
		if gw := gatewayMap[option.IngSvc]; gw != nil {
			// assuming 1 IP per gateway
			gwStatus := gw.Status.DeepCopy()
			for _, vip := range option.Vip {
				gwStatus.Addresses = []advl4v1alpha1pre1.GatewayAddress{{
					Value: vip,
					Type:  advl4v1alpha1pre1.IPAddressType,
				}}
			}

			// when statuses are synced during bootup
			InitializeGatewayConditions(gwStatus, &gw.Spec, true)
			UpdateGatewayStatusGWCondition(option.Key, gwStatus, &UpdateGWStatusConditionOptions{
				Type:   "Ready",
				Status: corev1.ConditionTrue,
			})
			UpdateGatewayStatusObject(option.Key, gw, gwStatus)
			skipDelete[option.IngSvc] = true
		}
	}

	// reset IPAddress and finalizer from Gateways that do not have a corresponding VS in cache
	if bulk {
		for gwNSName := range gatewayMap {
			if val, ok := skipDelete[gwNSName]; ok && val {
				continue
			}
			l.DeleteGatewayStatusAddress(lib.ServiceMetadataObj{
				Gateway: gwNSName,
			}, lib.SyncStatusKey)
		}
	}
}

func parseOptionsFromMetadata(options []UpdateOptions, bulk bool) ([]string, []UpdateOptions) {
	var objectsToUpdate []string
	var updateGWOptions []UpdateOptions

	for _, option := range options {
		if option.ServiceMetadata.Gateway != "" {
			option.IngSvc = option.ServiceMetadata.Gateway
			objectsToUpdate = append(objectsToUpdate, option.ServiceMetadata.Gateway)
			updateGWOptions = append(updateGWOptions, option)
		}
	}
	return objectsToUpdate, updateGWOptions
}

func (l *leader) DeleteGatewayStatusAddress(svcMetadataObj lib.ServiceMetadataObj, key string) error {
	gwNSName := strings.Split(svcMetadataObj.Gateway, "/")
	if utils.IsWCP() {
		gw, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(gwNSName[0]).Get(gwNSName[1])
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: there was a problem in resetting the gateway address status: %s", key, err)
			return err
		}

		if len(gw.Status.Addresses) == 0 || gw.Status.Addresses[0].Value == "" {
			return nil
		}

		// assuming 1 IP per gateway
		gwStatus := gw.Status.DeepCopy()
		gwStatus.Addresses = []advl4v1alpha1pre1.GatewayAddress{}
		UpdateGatewayStatusGWCondition(key, gwStatus, &UpdateGWStatusConditionOptions{
			Type:   "Pending",
			Status: corev1.ConditionTrue,
			Reason: "virtualservice deleted/notfound",
		})
		UpdateGatewayStatusObject(key, gw, gwStatus)
	} else if lib.UseServicesAPI() {
		return DeleteSvcApiStatus(key, svcMetadataObj)
	}
	utils.AviLog.Infof("key: %s, msg: Successfully reset the address status of gateway: %s", key, svcMetadataObj.Gateway)
	return nil
}

// supported GatewayConditionTypes
// InvalidListeners, InvalidAddress, *Serviceable
func UpdateGatewayStatusGWCondition(key string, gwStatus *advl4v1alpha1pre1.GatewayStatus, updateStatus *UpdateGWStatusConditionOptions) {
	utils.AviLog.Debugf("Updating Gateway status gateway condition %v", utils.Stringify(updateStatus))
	InitializeGatewayConditions(gwStatus, nil, false)

	for i := range gwStatus.Conditions {
		if string(gwStatus.Conditions[i].Type) == updateStatus.Type {
			gwStatus.Conditions[i].Status = updateStatus.Status
			gwStatus.Conditions[i].Message = updateStatus.Message
			gwStatus.Conditions[i].Reason = updateStatus.Reason
			gwStatus.Conditions[i].LastTransitionTime = metav1.Now()
		}

		var inverseCondition corev1.ConditionStatus
		if updateStatus.Status == corev1.ConditionFalse {
			inverseCondition = corev1.ConditionTrue
		} else {
			inverseCondition = corev1.ConditionFalse
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

	var listenerConditionStatus corev1.ConditionStatus
	if updateStatus.Type == "Ready" {
		listenerConditionStatus = corev1.ConditionTrue
	} else {
		listenerConditionStatus = corev1.ConditionFalse
	}
	UpdateGatewayStatusListenerConditions(key, gwStatus, "", &UpdateGWStatusConditionOptions{
		Type:   "Ready",
		Status: listenerConditionStatus,
	})
}

// supported ListenerConditionType
// PortConflict, InvalidRoutes, UnsupportedProtocol, *Serviceable
// pass portString as empty string for updating status in all ports
func UpdateGatewayStatusListenerConditions(key string, gwStatus *advl4v1alpha1pre1.GatewayStatus, portString string, updateStatus *UpdateGWStatusConditionOptions) {
	utils.AviLog.Debugf("key: %s, msg: Updating Gateway status listener condition port: %s %v", key, updateStatus, portString, utils.Stringify(updateStatus))
	for port, condition := range gwStatus.Listeners {
		notFound := true
		if condition.Port == portString || portString == "" {
			for i, portCondition := range condition.Conditions {
				if updateStatus.Type == "Ready" && updateStatus.Type != string(portCondition.Type) && updateStatus.Status == corev1.ConditionTrue {
					gwStatus.Listeners[port].Conditions[i].Status = corev1.ConditionFalse
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
				gwStatus.Listeners[port].Conditions = append(gwStatus.Listeners[port].Conditions, advl4v1alpha1pre1.ListenerCondition{
					Type:               advl4v1alpha1pre1.ListenerConditionType(updateStatus.Type),
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
		UpdateGatewayStatusGWCondition(key, gwStatus, &UpdateGWStatusConditionOptions{
			Type:   "Pending",
			Status: corev1.ConditionTrue,
			Reason: fmt.Sprintf("port %s error %s", portString, updateStatus.Type),
		})
	}
}

func UpdateGatewayStatusObject(key string, gw *advl4v1alpha1pre1.Gateway, updateStatus *advl4v1alpha1pre1.GatewayStatus, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 5 {
			utils.AviLog.Errorf("key: %s, msg: UpdateGatewayStatus retried 5 times, aborting", key)
			return
		}
	}

	// if an IP address is present on the gateway object, it is fair to assume that the gateway corresponds to a VS in avi
	// in case the IP is not present, the gateway can be deleted freely since deleting that would be a NOOP for AKO
	// so we add finalizer when an IP is updated to the gateway, and remove it when we delete the IP address.
	if len(updateStatus.Addresses) > 0 && updateStatus.Addresses[0].Value != "" {
		lib.CheckAndSetGatewayFinalizer(gw)
	} else {
		lib.RemoveGatewayFinalizer(gw)
	}

	if compareGatewayStatuses(&gw.Status, updateStatus) {
		return
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": updateStatus,
	})

	_, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: %d there was an error in updating the gateway status: %+v", key, retry, err)
		updatedGW, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(gw.Namespace).Get(gw.Name)
		if err != nil {
			utils.AviLog.Warnf("key: %s, gateway not found %v", key, err)
			return
		}
		UpdateGatewayStatusObject(key, updatedGW, updateStatus, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the gateway %s/%s status %+v", key, gw.Namespace, gw.Name, utils.Stringify(updateStatus))
}

func InitializeGatewayConditions(gwStatus *advl4v1alpha1pre1.GatewayStatus, gwSpec *advl4v1alpha1pre1.GatewaySpec, gwReady bool) {
	if len(gwStatus.Conditions) == 0 {
		gwStatus.Conditions = []advl4v1alpha1pre1.GatewayCondition{{
			Type:               "Pending",
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
		}, {
			Type:               "Ready",
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
		}}
	}

	if gwSpec == nil {
		return
	}

	gwPortMap := make(map[string][]advl4v1alpha1pre1.ListenerCondition)
	for _, listenerStatus := range gwStatus.Listeners {
		gwPortMap[listenerStatus.Port] = listenerStatus.Conditions
	}

	var listenerStatuses []advl4v1alpha1pre1.ListenerStatus
	for _, listener := range gwSpec.Listeners {
		if val, ok := gwPortMap[strconv.Itoa(int(listener.Port))]; ok {
			listenerStatuses = append(listenerStatuses, advl4v1alpha1pre1.ListenerStatus{
				Port:       strconv.Itoa(int(listener.Port)),
				Conditions: val,
			})
		} else {
			var portCondition corev1.ConditionStatus
			if gwReady {
				portCondition = corev1.ConditionTrue
			} else {
				portCondition = corev1.ConditionFalse
			}
			listenerStatuses = append(listenerStatuses, advl4v1alpha1pre1.ListenerStatus{
				Port: strconv.Itoa(int(listener.Port)),
				Conditions: []advl4v1alpha1pre1.ListenerCondition{{
					Type:               "Ready",
					Status:             portCondition,
					LastTransitionTime: metav1.Now(),
				}},
			})
		}
	}

	gwStatus.Listeners = listenerStatuses
	if len(gwStatus.Addresses) == 0 {
		gwStatus.Addresses = []advl4v1alpha1pre1.GatewayAddress{}
	}
	return
}

// getGateways fetches all ingresses and returns a map: {"namespace/name": ingressObj...}
// if bulk is set to true, this fetches all ingresses in a single k8s api-server call
func getGateways(gwNSNames []string, bulk bool, retryNum ...int) map[string]*advl4v1alpha1pre1.Gateway {
	retry := 0
	gwMap := make(map[string]*advl4v1alpha1pre1.Gateway)
	if len(retryNum) > 0 {
		utils.AviLog.Infof("Retrying to get the gateway for status update")
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("getGateways for status update retried 3 times, aborting")
			return gwMap
		}
	}

	if bulk {
		// Get GatewayClasses with Avi set as the controller, get corresponding Gateways,
		// to return all AKO ingestable Gateways.
		aviGWClasses := make(map[string]bool)
		gwClassList, err := lib.AKOControlConfig().AdvL4Informers().GatewayClassInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("Could not get the GatewayClass object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getGateways(gwNSNames, bulk, retry+1)
			}
		}

		if len(gwClassList) == 0 {
			return gwMap
		}

		for i := range gwClassList {
			if gwClassList[i].Spec.Controller == lib.AviGatewayController {
				aviGWClasses[gwClassList[i].Name] = true
			}
		}

		gwList, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("Could not get the Gateway object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getGateways(gwNSNames, bulk, retry+1)
			}
		}

		for i := range gwList {
			if _, ok := aviGWClasses[gwList[i].Spec.Class]; ok {
				gw := gwList[i]
				if utils.CheckIfNamespaceAccepted(gw.Namespace) {
					gwMap[gw.Namespace+"/"+gw.Name] = gw
				}
			}
		}

		return gwMap
	}

	for _, namespaceName := range gwNSNames {
		nsNameSplit := strings.Split(namespaceName, "/")
		gw, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(nsNameSplit[0]).Get(nsNameSplit[1])
		if err != nil {
			utils.AviLog.Warnf("Could not get the gateway object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getGateways(gwNSNames, bulk, retry+1)
			}
		} else {
			gwMap[gw.Namespace+"/"+gw.Name] = gw
		}

	}

	return gwMap
}

// do not compare lastTransitionTime updates in gateway
func compareGatewayStatuses(old, new *advl4v1alpha1pre1.GatewayStatus) bool {
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

func (f *follower) UpdateGatewayStatusAddress(options []UpdateOptions, bulk bool) {
	for _, option := range options {
		utils.AviLog.Debugf("key: %s, AKO is not a leader, not updating the L4 LB status", option.Key)
	}
}

func (f *follower) DeleteGatewayStatusAddress(svcMetadataObj lib.ServiceMetadataObj, key string) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not deleting the L4 LB status", key)
	return nil
}

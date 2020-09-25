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
	"errors"
	"fmt"
	"strconv"
	"strings"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpdateGWStatusConditionOptions struct {
	Type    string               // to be casted to the appropriate conditionType
	Status  core.ConditionStatus // True/False/Unknown
	Message string               // extended condition message
	Reason  string               // reason for transition
}

// TODO: handle bulk during bootup
func UpdateGatewayStatusAddress(options []UpdateStatusOptions, bulk bool) {
	for _, option := range options {
		gatewayNSName := strings.Split(option.ServiceMetadata.Gateway, "/")
		gw, err := lib.GetAdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gatewayNSName[0]).Get(gatewayNSName[1], metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Infof("key: %s, msg: unable to find gateway object %s", option.Key, option.ServiceMetadata.Gateway)
			continue
		}

		// assuming 1 IP per gateway
		gwStatus := gw.Status
		if len(gwStatus.Addresses) > 0 && gwStatus.Addresses[0].Value == option.Vip {
			continue
		}

		gwStatus.Addresses = []advl4v1alpha1pre1.GatewayAddress{{
			Value: option.Vip,
			Type:  advl4v1alpha1pre1.IPAddressType,
		}}
		UpdateGatewayStatusGWCondition(gw, &UpdateGWStatusConditionOptions{
			Type:   "Ready",
			Status: corev1.ConditionTrue,
		})
		UpdateGatewayStatusObject(gw, &gwStatus)

		utils.AviLog.Debugf("key: %s, msg: Updating corresponding service %v statuses for gateway %s",
			option.Key, option.ServiceMetadata.NamespaceServiceName, option.ServiceMetadata.Gateway)

		UpdateL4LBStatus([]UpdateStatusOptions{{
			Vip: option.Vip,
			Key: option.Key,
			ServiceMetadata: avicache.ServiceMetadataObj{
				NamespaceServiceName: option.ServiceMetadata.NamespaceServiceName,
			},
		}}, false)
	}

	return
}

func DeleteGatewayStatusAddress(svcMetadataObj avicache.ServiceMetadataObj, key string) error {
	gwNSName := strings.Split(svcMetadataObj.Gateway, "/")
	gw, err := lib.GetAdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gwNSName[0]).Get(gwNSName[1], metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was a problem in resetting the gateway address status: %s", key, err)
		return err
	}

	if len(gw.Status.Addresses) == 0 ||
		(len(gw.Status.Addresses) > 0 && gw.Status.Addresses[0].Value == "") {
		return nil
	}

	// assuming 1 IP per gateway
	gw.Status.Addresses = []advl4v1alpha1pre1.GatewayAddress{}
	UpdateGatewayStatusGWCondition(gw, &UpdateGWStatusConditionOptions{
		Type:   "Pending",
		Status: corev1.ConditionTrue,
		Reason: "virtualservice deleted/notfound",
	})
	UpdateGatewayStatusObject(gw, &gw.Status)

	utils.AviLog.Infof("key: %s, msg: Successfully reset the address status of gateway: %s", key, svcMetadataObj.Gateway)
	return nil
}

// supported GatewayConditionTypes
// InvalidListeners, InvalidAddress, *Serviceable
func UpdateGatewayStatusGWCondition(gw *advl4v1alpha1pre1.Gateway, updateStatus *UpdateGWStatusConditionOptions) {
	utils.AviLog.Debugf("Updating Gateway status gateway condition %v", utils.Stringify(updateStatus))
	for i, _ := range gw.Status.Conditions {
		if string(gw.Status.Conditions[i].Type) == updateStatus.Type {
			gw.Status.Conditions[i].Status = updateStatus.Status
			gw.Status.Conditions[i].Message = updateStatus.Message
			gw.Status.Conditions[i].Reason = updateStatus.Reason
			gw.Status.Conditions[i].LastTransitionTime = metav1.Now()
		}

		if (updateStatus.Type == "Pending" && string(gw.Status.Conditions[i].Type) == "Ready") ||
			(updateStatus.Type == "Ready" && string(gw.Status.Conditions[i].Type) == "Pending") {
			// if Pending true, mark Ready as false automatically
			// if Ready true, mark Pending as false automatically
			gw.Status.Conditions[i].Status = corev1.ConditionFalse
			gw.Status.Conditions[i].LastTransitionTime = metav1.Now()
			gw.Status.Conditions[i].Message = ""
			gw.Status.Conditions[i].Reason = ""
		}

		if updateStatus.Type == "Ready" {
			UpdateGatewayStatusListenerConditions(gw, "", &UpdateGWStatusConditionOptions{
				Type:   "Ready",
				Status: corev1.ConditionTrue,
			})
		}
	}
}

// supported ListenerConditionType
// PortConflict, InvalidRoutes, UnsupportedProtocol, *Serviceable
func UpdateGatewayStatusListenerConditions(gw *advl4v1alpha1pre1.Gateway, portString string, updateStatus *UpdateGWStatusConditionOptions) {
	utils.AviLog.Debugf("Updating Gateway status listener condition port: %s %v", portString, utils.Stringify(updateStatus))
	for port, condition := range gw.Status.Listeners {
		notFound := true
		if condition.Port == portString || portString == "" {
			for i, portCondition := range condition.Conditions {
				if updateStatus.Type == "Ready" && updateStatus.Type != string(portCondition.Type) {
					gw.Status.Listeners[port].Conditions[i].Status = corev1.ConditionFalse
					gw.Status.Listeners[port].Conditions[i].Message = ""
					gw.Status.Listeners[port].Conditions[i].Reason = ""
				}

				if string(portCondition.Type) == updateStatus.Type {
					gw.Status.Listeners[port].Conditions[i].Status = updateStatus.Status
					gw.Status.Listeners[port].Conditions[i].Message = updateStatus.Message
					gw.Status.Listeners[port].Conditions[i].Reason = updateStatus.Reason
					gw.Status.Listeners[port].Conditions[i].LastTransitionTime = metav1.Now()
					notFound = false
				}
			}

			if notFound {
				gw.Status.Listeners[port].Conditions = append(gw.Status.Listeners[port].Conditions, advl4v1alpha1pre1.ListenerCondition{
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
		UpdateGatewayStatusGWCondition(gw, &UpdateGWStatusConditionOptions{
			Type:   "Pending",
			Status: corev1.ConditionTrue,
			Reason: fmt.Sprintf("port %s error %s", portString, updateStatus.Type),
		})
		UpdateGatewayStatusGWCondition(gw, &UpdateGWStatusConditionOptions{
			Type:   "Ready",
			Status: corev1.ConditionFalse,
			Reason: "NotReady",
		})
	}
}

func UpdateGatewayStatusObject(gw *advl4v1alpha1pre1.Gateway, updateStatus *advl4v1alpha1pre1.GatewayStatus, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 4 {
			return errors.New("msg: UpdateGatewayStatus retried 5 times, aborting")
		}
	}

	gw.Status = *updateStatus
	_, err := lib.GetAdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gw.Namespace).UpdateStatus(gw)
	if err != nil {
		utils.AviLog.Warnf("msg: %d there was an error in updating the gateway status: %+v", retry, err)
		updatedGW, err := lib.GetAdvL4Clientset().NetworkingV1alpha1pre1().Gateways(gw.Namespace).Get(gw.Name, metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("gateway not found %v", err)
			return err
		}
		return UpdateGatewayStatusObject(updatedGW, updateStatus, retry+1)
	}

	utils.AviLog.Infof("msg: Successfully updated the gateway %s/%s status %+v", gw.Namespace, gw.Name, utils.Stringify(updateStatus))
	return nil
}

func InitializeGatewayConditions(gw *advl4v1alpha1pre1.Gateway) error {
	if len(gw.Status.Conditions) > 0 {
		// already initialised
		return nil
	}

	gw.Status.Conditions = []advl4v1alpha1pre1.GatewayCondition{{
		Type:               "Pending",
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	}, {
		Type:               "Ready",
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}}

	var listenerStatuses []advl4v1alpha1pre1.ListenerStatus
	for _, listener := range gw.Spec.Listeners {
		listenerStatuses = append(listenerStatuses, advl4v1alpha1pre1.ListenerStatus{
			Port: strconv.Itoa(int(listener.Port)),
			Conditions: []advl4v1alpha1pre1.ListenerCondition{{
				Type:               "Ready",
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
			}},
		})
	}
	gw.Status.Listeners = listenerStatuses
	gw.Status.Addresses = []advl4v1alpha1pre1.GatewayAddress{}

	return UpdateGatewayStatusObject(gw, &gw.Status)
}

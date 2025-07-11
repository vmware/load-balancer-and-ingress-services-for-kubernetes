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
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
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

type gateway struct{}

func (o *gateway) Get(key string, option status.StatusOptions) *gatewayv1.Gateway {

	nsName := strings.Split(option.Options.ServiceMetadata.Gateway, "/")
	if len(nsName) != 2 {
		utils.AviLog.Warnf("key: %s, msg: invalid gateway name and namespace", key)
		return nil
	}
	namespace := nsName[0]
	name := nsName[1]

	gw, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the gateway object. err: %s", key, err)
		return nil
	}
	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the gateway object %s", key, name)
	return gw.DeepCopy()
}

func (o *gateway) GetAll(key string) map[string]*gatewayv1.Gateway {

	gwClassList, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the gateway class objects. err: %s", key, err)
		return nil
	}

	if len(gwClassList) == 0 {
		return nil
	}

	gwClassOwnedByAko := make(map[string]struct{})
	for i := range gwClassList {
		if gwClassList[i].Spec.ControllerName == akogatewayapilib.GatewayController {
			gwClassOwnedByAko[gwClassList[i].Name] = struct{}{}
		}
	}

	gwList, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the gateway objects owned by AKO. err: %s", key, err)
		return nil
	}

	gwMap := make(map[string]*gatewayv1.Gateway)
	for _, gw := range gwList {
		if _, ok := gwClassOwnedByAko[string(gw.Spec.GatewayClassName)]; ok {
			gwMap[gw.Namespace+"/"+gw.Name] = gw.DeepCopy()
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the gateway objects owned by AKO", key)
	return gwMap
}

func (o *gateway) Delete(key string, option status.StatusOptions) {

	gw := o.Get(key, option)
	if gw == nil {
		return
	}

	// Gateway don't have any address. In this case, the delete is not required.
	if len(gw.Status.Addresses) == 0 ||
		(len(gw.Status.Addresses) > 0 && gw.Status.Addresses[0].Value == "") {
		return
	}

	// assuming 1 IP per gateway
	gatewayStatus := gw.Status.DeepCopy()
	namespace := gw.Namespace
	name := gw.Name

	gwStatusFromMapping := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(namespace + "/" + name)
	if gwStatusFromMapping != nil {
		gatewayStatus = gwStatusFromMapping
	}
	gatewayStatus.Addresses = []gatewayv1.GatewayStatusAddress{}

	condition := NewCondition()
	condition.
		Type(string(gatewayv1.GatewayConditionProgrammed)).
		Status(metav1.ConditionUnknown).
		Reason(string(gatewayv1.GatewayReasonPending)).
		ObservedGeneration(gw.ObjectMeta.Generation).
		Message("Virtual service has been deleted").
		SetIn(&gatewayStatus.Conditions)

	for i := range gatewayStatus.Listeners {
		listenerCondition := NewCondition()
		listenerCondition.
			Type(string(gatewayv1.GatewayConditionProgrammed)).
			Status(metav1.ConditionUnknown).
			Reason(string(gatewayv1.GatewayReasonPending)).
			ObservedGeneration(gw.ObjectMeta.Generation).
			Message("Virtual service has been deleted").
			SetIn(&gatewayStatus.Listeners[i].Conditions)
	}

	o.Patch(key, gw, &status.Status{GatewayStatus: gatewayStatus})
	utils.AviLog.Infof("key: %s, msg: Successfully reset the address status of gateway: %s", key, gw.Name)

	// TODO: Add annotation delete code here
}

func (o *gateway) Update(key string, option status.StatusOptions) {
	gw := o.Get(key, option)
	if gw == nil {
		return
	}

	if option.Options != nil && option.Options.Status != nil && option.Options.Status.GatewayStatus != nil {
		o.Patch(key, gw, option.Options.Status)
		return
	}

	gatewaystatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gw.Namespace + "/" + gw.Name)
	addressType := gatewayv1.IPAddressType
	ipAddrs := []gatewayv1.GatewayStatusAddress{}
	for _, vip := range option.Options.Vip {
		ipAddrs = append(ipAddrs, gatewayv1.GatewayStatusAddress{
			Type:  &addressType,
			Value: vip,
		})
	}
	if gatewaystatus == nil {
		gatewaystatus = gw.Status.DeepCopy()
	}
	gatewaystatus.Addresses = ipAddrs

	condition := NewCondition()
	var conditionType, reason, message string
	conditionStatus := metav1.ConditionTrue

	if option.Options.Message != "" {
		conditionType = string(gatewayv1.GatewayConditionProgrammed)
		conditionStatus = metav1.ConditionFalse
		reason = string(gatewayv1.GatewayReasonInvalid)
		message = option.Options.Message
	} else {
		conditionType = string(gatewayv1.GatewayConditionProgrammed)
		reason = string(gatewayv1.GatewayReasonProgrammed)
		message = "Virtual service configured/updated"
	}
	condition.
		Type(conditionType).
		Status(conditionStatus).
		Reason(reason).
		ObservedGeneration(gw.ObjectMeta.Generation).
		Message(message).
		SetIn(&gatewaystatus.Conditions)

	for i, listener := range gatewaystatus.Listeners {
		listenerCondition := NewCondition()
		if listener.Conditions[0].Type == string(gatewayv1.ListenerConditionAccepted) && listener.Conditions[0].Status == metav1.ConditionTrue && listener.Conditions[1].Type == string(gatewayv1.ListenerConditionResolvedRefs) && listener.Conditions[1].Status == metav1.ConditionTrue {
			listenerCondition.
				Type(string(gatewayv1.ListenerConditionProgrammed)).
				Status(conditionStatus).
				Reason(string(gatewayv1.ListenerReasonProgrammed)).
				ObservedGeneration(gw.ObjectMeta.Generation).
				Message(message).
				SetIn(&gatewaystatus.Listeners[i].Conditions)
		} else {
			listenerCondition.
				Type(string(gatewayv1.ListenerConditionProgrammed)).
				Status(metav1.ConditionFalse).
				Reason(string(gatewayv1.ListenerReasonInvalid)).
				ObservedGeneration(gw.ObjectMeta.Generation).
				Message("Virtual service not configured/updated for this listener").
				SetIn(&gatewaystatus.Listeners[i].Conditions)
		}
	}
	o.Patch(key, gw, &status.Status{GatewayStatus: gatewaystatus})

	// TODO: Annotation update code here
}

func (o *gateway) BulkUpdate(key string, options []status.StatusOptions) {

	gwMap := o.GetAll(key)
	for _, option := range options {
		nsName := option.Options.ServiceMetadata.Gateway
		if gw, ok := gwMap[nsName]; ok {
			gatewaystatus := &gatewayv1.GatewayStatus{}
			addressType := gatewayv1.IPAddressType
			gatewaystatus.Addresses = append(gatewaystatus.Addresses, gatewayv1.GatewayStatusAddress{
				Type:  &addressType,
				Value: option.Options.Vip[0],
			})
			apimeta.SetStatusCondition(&gatewaystatus.Conditions, metav1.Condition{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionTrue,
				Reason:             string(gatewayv1.GatewayReasonProgrammed),
				Message:            "Virtual service configured/updated",
				ObservedGeneration: gw.ObjectMeta.Generation + 1,
			})

			o.Patch(key, gw, &status.Status{GatewayStatus: gatewaystatus})

			// TODO: Annotation update code here
		}
	}
}

func (o *gateway) Patch(key string, obj runtime.Object, status *status.Status, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 5 {
			utils.AviLog.Errorf("key: %s, msg: Patch retried 5 times, aborting", key)
			akogatewayapilib.AKOControlConfig().EventRecorder().Eventf(obj, corev1.EventTypeWarning, lib.PatchFailed, "Patch of status failed after multiple retries")
			return errors.New("Patch retried 5 times, aborting")
		}
	}

	gw := obj.(*gatewayv1.Gateway)
	if o.isStatusEqual(&gw.Status, status.GatewayStatus) {
		return nil
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": status.GatewayStatus,
	})
	_, err := akogatewayapilib.AKOControlConfig().GatewayAPIClientset().GatewayV1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was an error in updating the gateway status. err: %+v, retry: %d", key, err, retry)
		updatedGW, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(gw.Namespace).Get(gw.Name)
		if err != nil {
			utils.AviLog.Warnf("gateway not found %v", err)
			return err
		}
		return o.Patch(key, updatedGW, status, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the gateway %s/%s status %+v", key, gw.Namespace, gw.Name, utils.Stringify(status))
	return nil
}

func (o *gateway) isStatusEqual(old, new *gatewayv1.GatewayStatus) bool {
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

/*
 * Copyright 2023-2024 VMware, Inc.
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
	"reflect"
	"strings"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type gateway struct{}

func (o *gateway) Get(key string, option status.StatusOptions) *gatewayv1beta1.Gateway {

	nsName := strings.Split(option.Options.ServiceMetadata.Gateway, "/")
	if len(nsName) != 2 {
		utils.AviLog.Warnf("key: %s, msg: invalid gateway name and namespace", key)
		return nil
	}
	namespace := nsName[0]
	name := nsName[1]

	gw, err := lib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the gateway object. err: %s", key, err)
		return nil
	}
	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the gateway object %s", key, name)
	return gw.DeepCopy()
}

func (o *gateway) GetAll(key string) map[string]*gatewayv1beta1.Gateway {

	gwClassList, err := lib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the gateway class objects. err: %s", key, err)
		return nil
	}

	if len(gwClassList) == 0 {
		return nil
	}

	gwClassOwnedByAko := make(map[string]struct{})
	for i := range gwClassList {
		if gwClassList[i].Spec.ControllerName == "ako.vmware.com/avi-lb" { // TODO: re-use the constant
			gwClassOwnedByAko[gwClassList[i].Name] = struct{}{}
		}
	}

	gwList, err := lib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the gateway objects owned by AKO. err: %s", key, err)
		return nil
	}

	gwMap := make(map[string]*gatewayv1beta1.Gateway)
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
	status := gw.Status.DeepCopy()
	status.Addresses = []gatewayv1beta1.GatewayAddress{}

	condition := NewCondition()
	condition.
		Type(string(gatewayv1beta1.GatewayConditionProgrammed)).
		Status(metav1.ConditionUnknown).
		Reason(string(gatewayv1beta1.GatewayReasonPending)).
		ObservedGeneration(gw.ObjectMeta.Generation).
		Message("Virtual service has been deleted").
		SetIn(&status.Conditions)

	for i := range status.Listeners {
		listenerCondition := NewCondition()
		listenerCondition.
			Type(string(gatewayv1beta1.GatewayConditionProgrammed)).
			Status(metav1.ConditionUnknown).
			Reason(string(gatewayv1beta1.GatewayReasonPending)).
			ObservedGeneration(gw.ObjectMeta.Generation).
			Message("Virtual service has been deleted").
			SetIn(&status.Listeners[i].Conditions)
	}

	o.Patch(key, gw, &Status{*status})
	utils.AviLog.Infof("key: %s, msg: Successfully reset the address status of gateway: %s", key, gw.Name)

	// TODO: Add annotation delete code here
}

func (o *gateway) Update(key string, option status.StatusOptions) {
	gw := o.Get(key, option)
	if gw == nil {
		return
	}

	status := gw.Status.DeepCopy()
	addressType := gatewayv1beta1.IPAddressType
	status.Addresses = append(status.Addresses, gatewayv1beta1.GatewayAddress{
		Type:  &addressType,
		Value: option.Options.Vip[0],
	})

	// TODO: Add a way to propagate the error from the Rest layer to status layer.

	condition := NewCondition()
	condition.
		Type(string(gatewayv1beta1.GatewayConditionProgrammed)).
		Status(metav1.ConditionTrue).
		Reason(string(gatewayv1beta1.GatewayReasonProgrammed)).
		ObservedGeneration(gw.ObjectMeta.Generation).
		Message("Virtual service configured/updated").
		SetIn(&status.Conditions)

	for i := range status.Listeners {
		listenerCondition := NewCondition()
		listenerCondition.
			Type(string(gatewayv1beta1.GatewayConditionProgrammed)).
			Status(metav1.ConditionTrue).
			Reason(string(gatewayv1beta1.GatewayReasonProgrammed)).
			ObservedGeneration(gw.ObjectMeta.Generation).
			SetIn(&status.Listeners[i].Conditions)
	}

	o.Patch(key, gw, &Status{*status})

	// TODO: Annotation update code here
}

func (o *gateway) BulkUpdate(key string, options []status.StatusOptions) {

	gwMap := o.GetAll(key)
	for _, option := range options {
		nsName := option.Options.ServiceMetadata.Gateway
		if gw, ok := gwMap[nsName]; ok {
			status := gatewayv1beta1.GatewayStatus{}
			addressType := gatewayv1beta1.IPAddressType
			status.Addresses = append(status.Addresses, gatewayv1beta1.GatewayAddress{
				Type:  &addressType,
				Value: option.Options.Vip[0],
			})
			apimeta.SetStatusCondition(&status.Conditions, metav1.Condition{
				Type:               string(gatewayv1beta1.GatewayConditionProgrammed),
				Status:             metav1.ConditionTrue,
				Reason:             string(gatewayv1beta1.GatewayReasonProgrammed),
				Message:            "Virtual service configured/updated",
				ObservedGeneration: gw.ObjectMeta.Generation + 1,
			})
			o.Patch(key, gw, &Status{status})

			// TODO: Annotation update code here
		}
	}
}

func (o *gateway) Patch(key string, obj runtime.Object, status *Status, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 5 {
			utils.AviLog.Errorf("key: %s, msg: Patch retried 5 times, aborting", key)
			return
		}
	}

	gw := obj.(*gatewayv1beta1.Gateway)
	if o.isGatewayStatusEqual(&gw.Status, &status.GatewayStatus) {
		return
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": status,
	})
	_, err := lib.AKOControlConfig().GatewayAPIClientset().GatewayV1beta1().Gateways(gw.Namespace).Patch(context.TODO(), gw.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was an error in updating the gateway status. err: %+v, retry: %d", key, err, retry)
		updatedGW, err := lib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(gw.Namespace).Get(gw.Name)
		if err != nil {
			utils.AviLog.Warnf("gateway not found %v", err)
			return
		}
		o.Patch(key, updatedGW, status, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the gateway %s/%s status %+v", key, gw.Namespace, gw.Name, utils.Stringify(status))
}

func (o *gateway) isGatewayStatusEqual(old, new *gatewayv1beta1.GatewayStatus) bool {
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

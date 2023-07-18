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

package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func IsValidGateway(key string, gateway *gatewayv1beta1.Gateway) bool {
	spec := gateway.Spec

	// is associated with gateway class
	if CheckGatewayController(*gateway) {
		utils.AviLog.Warnf("key %s, msg: controller is not ako, ignoring the gateway object %s", key, gateway.Name)
		return false
	}

	defaultCondition := status.NewCondition().
		Type(string(gatewayv1beta1.GatewayConditionAccepted)).
		Reason(string(gatewayv1beta1.GatewayReasonInvalid)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)

	gatewayStatus := gateway.Status.DeepCopy()

	// has 1 or more listeners
	if len(spec.Listeners) == 0 {
		utils.AviLog.Errorf("key %s, msg: no listeners found in gateway %+v", key, gateway.Name)
		defaultCondition.
			Message("No listeners found").
			SetIn(&gatewayStatus.Conditions)
		status.Record(key, gateway, &status.Status{GatewayStatus: *gatewayStatus})
		return false
	}

	// has 1 or none addresses
	if len(spec.Addresses) > 1 {
		utils.AviLog.Errorf("key: %s, msg: more than 1 gateway address found in gateway %+v", key, gateway.Name)
		defaultCondition.
			Message("More than one address is not supported").
			SetIn(&gatewayStatus.Conditions)
		status.Record(key, gateway, &status.Status{GatewayStatus: *gatewayStatus})
		return false
	}
	if len(spec.Addresses) == 1 && *spec.Addresses[0].Type != "IPAddress" {
		utils.AviLog.Errorf("gateway address is not of type IPAddress %+v", gateway.Name)
		defaultCondition.
			Message("Only IPAddress as AddressType is supported").
			SetIn(&gatewayStatus.Conditions)
		status.Record(key, gateway, &status.Status{GatewayStatus: *gatewayStatus})
		return false
	}

	gatewayStatus.Listeners = make([]gatewayv1beta1.ListenerStatus, len(gateway.Spec.Listeners))

	var invalidListenerCount int
	for index := range spec.Listeners {
		if !isValidListener(key, gateway, gatewayStatus, index) {
			invalidListenerCount++
		}
	}

	if invalidListenerCount > 0 {
		utils.AviLog.Errorf("key: %s, msg: Gateway %s contains %d invalid listeners", key, gateway.Name, invalidListenerCount)
		defaultCondition.
			Type(string(gatewayv1beta1.GatewayReasonAccepted)).
			Reason(string(gatewayv1beta1.GatewayReasonListenersNotValid)).
			Message(fmt.Sprintf("Gateway contains %d invalid listener(s)", invalidListenerCount)).
			SetIn(&gatewayStatus.Conditions)
		status.Record(key, gateway, &status.Status{GatewayStatus: *gatewayStatus})
		return false
	}

	defaultCondition.
		Reason(string(gatewayv1beta1.GatewayReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Gateway configuration is valid").
		SetIn(&gatewayStatus.Conditions)
	status.Record(key, gateway, &status.Status{GatewayStatus: *gatewayStatus})
	utils.AviLog.Infof("key: %s, msg: Gateway %s is valid", key, gateway.Name)
	return true
}

func isValidListener(key string, gateway *gatewayv1beta1.Gateway, gatewayStatus *gatewayv1beta1.GatewayStatus, index int) bool {

	listener := gateway.Spec.Listeners[index]
	gatewayStatus.Listeners[index].Name = gateway.Spec.Listeners[index].Name
	gatewayStatus.Listeners[index].SupportedKinds = akogatewayapilib.SupportedKinds[listener.Protocol]
	gatewayStatus.Listeners[index].AttachedRoutes = akogatewayapilib.ZeroAttachedRoutes

	defaultCondition := status.NewCondition().
		Type(string(gatewayv1beta1.GatewayConditionAccepted)).
		Reason(string(gatewayv1beta1.GatewayReasonListenersNotValid)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)

	// hostname is not nil or wildcard
	if listener.Hostname == nil || *listener.Hostname == "*" {
		utils.AviLog.Errorf("key: %s, msg: hostname with wildcard found in listener %s", key, listener.Name)
		defaultCondition.
			Message("Hostname not found or Hostname has invalid configuration").
			SetIn(&gatewayStatus.Listeners[index].Conditions)
		return false
	}

	// protocol validation
	if listener.Protocol != gatewayv1beta1.HTTPProtocolType &&
		listener.Protocol != gatewayv1beta1.HTTPSProtocolType {
		utils.AviLog.Errorf("key: %s, msg: protocol is not supported for listener %s", key, listener.Name)
		defaultCondition.
			Reason(string(gatewayv1beta1.ListenerReasonUnsupportedProtocol)).
			Message("Unsupported protocol").
			SetIn(&gatewayStatus.Listeners[index].Conditions)
		gatewayStatus.Listeners[index].SupportedKinds = akogatewayapilib.SupportedKinds[gatewayv1beta1.HTTPSProtocolType]
		return false
	}

	// has valid TLS config
	if listener.TLS != nil {
		if (listener.TLS.Mode != nil && *listener.TLS.Mode != gatewayv1beta1.TLSModeTerminate) || len(listener.TLS.CertificateRefs) == 0 {
			utils.AviLog.Errorf("key: %s, msg: tls mode/ref not valid %+v/%+v", key, gateway.Name, listener.Name)
			defaultCondition.
				Reason(string(gatewayv1beta1.ListenerReasonInvalidCertificateRef)).
				Message("TLS mode or reference not valid").
				SetIn(&gatewayStatus.Listeners[index].Conditions)
			return false
		}
		for _, certRef := range listener.TLS.CertificateRefs {
			//only secret is allowed
			if (certRef.Group != nil && string(*certRef.Group) != "") ||
				certRef.Kind != nil && string(*certRef.Kind) != utils.Secret {
				utils.AviLog.Errorf("CertificateRef is not valid %+v/%+v, must be Secret", gateway.Name, listener.Name)
				defaultCondition.
					Reason(string(gatewayv1beta1.ListenerReasonInvalidCertificateRef)).
					Message("TLS mode or reference not valid").
					SetIn(&gatewayStatus.Listeners[index].Conditions)
				return false
			}

		}
	}

	// Valid listener
	defaultCondition.
		Reason(string(gatewayv1beta1.GatewayReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Listener is valid").
		SetIn(&gatewayStatus.Listeners[index].Conditions)
	utils.AviLog.Infof("key: %s, msg: Listener %s/%s is valid", key, gateway.Name, listener.Name)
	return true
}

func CheckGatewayClassController(gwClass gatewayv1beta1.GatewayClass) bool {
	return gwClass.Spec.ControllerName == lib.AviIngressController
}

func CheckGatewayController(gw gatewayv1beta1.Gateway) bool {
	gwClass := string(gw.Spec.GatewayClassName)
	return objects.GatewayApiLister().IsGatewayClassPresent(gwClass)
}

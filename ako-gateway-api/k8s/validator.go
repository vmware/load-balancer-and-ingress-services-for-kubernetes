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
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapistatus "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func IsGatewayClassValid(key string, gatewayClass *gatewayv1.GatewayClass) bool {

	controllerName := string(gatewayClass.Spec.ControllerName)
	if !akogatewayapilib.CheckGatewayClassController(controllerName) {
		utils.AviLog.Errorf("key: %s, msg: Gateway controller is not AKO for GatewayClass object %s", key, gatewayClass.Name)
		return false
	}

	gatewayClassStatus := gatewayClass.Status.DeepCopy()
	akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.GatewayClassConditionStatusAccepted)).
		Reason(string(gatewayv1.GatewayClassReasonAccepted)).
		Status(metav1.ConditionTrue).
		ObservedGeneration(gatewayClass.ObjectMeta.Generation).
		Message("GatewayClass is valid").
		SetIn(&gatewayClassStatus.Conditions)
	akogatewayapistatus.Record(key, gatewayClass, &status.Status{GatewayClassStatus: gatewayClassStatus})
	utils.AviLog.Infof("key: %s, msg: GatewayClass object %s is valid", key, gatewayClass.Name)
	return true
}

func IsValidGateway(key string, gateway *gatewayv1.Gateway) (bool, bool) {
	spec := gateway.Spec
	allowedRoutesAll := false
	defaultCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.GatewayConditionAccepted)).
		Reason(string(gatewayv1.GatewayReasonInvalid)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)
	programmedCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.GatewayConditionProgrammed)).
		Reason(string(gatewayv1.GatewayReasonInvalid)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation).
		Message("Gateway not programmed")

	gatewayStatus := gateway.Status.DeepCopy()

	// has 1 or more listeners
	if len(spec.Listeners) == 0 {
		utils.AviLog.Errorf("key: %s, msg: no listeners found in gateway %+v", key, gateway.Name)
		defaultCondition.
			Message("No listeners found").
			SetIn(&gatewayStatus.Conditions)
		programmedCondition.
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &status.Status{GatewayStatus: gatewayStatus})
		return false, allowedRoutesAll
	}

	// has 1 or none addresses
	if len(spec.Addresses) > 1 {
		utils.AviLog.Errorf("key: %s, msg: more than 1 gateway address found in gateway %+v", key, gateway.Name)
		defaultCondition.
			Message("More than one address is not supported").
			SetIn(&gatewayStatus.Conditions)
		programmedCondition.
			Reason(string(gatewayv1.GatewayReasonAddressNotUsable)).
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &status.Status{GatewayStatus: gatewayStatus})
		return false, allowedRoutesAll
	}

	if len(spec.Addresses) == 1 && *spec.Addresses[0].Type != "IPAddress" {
		utils.AviLog.Errorf("key: %s, msg: gateway address is not of type IPAddress %+v", key, gateway.Name)
		defaultCondition.
			Reason(string(gatewayv1.GatewayReasonUnsupportedAddress)).
			Message("Only IPAddress as AddressType is supported").
			SetIn(&gatewayStatus.Conditions)
		programmedCondition.
			Reason(string(gatewayv1.GatewayReasonAddressNotUsable)).
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &status.Status{GatewayStatus: gatewayStatus})
		return false, allowedRoutesAll
	}

	gatewayStatus.Listeners = make([]gatewayv1.ListenerStatus, len(gateway.Spec.Listeners))

	var validListenerCount int
	for index := range spec.Listeners {
		if isValidListener(key, gateway, gatewayStatus, index) {
			if !allowedRoutesAll {
				if spec.Listeners[index].AllowedRoutes != nil && spec.Listeners[index].AllowedRoutes.Namespaces != nil && spec.Listeners[index].AllowedRoutes.Namespaces.From != nil {
					if string(*spec.Listeners[index].AllowedRoutes.Namespaces.From) == akogatewayapilib.AllowedRoutesNamespaceFromAll {
						allowedRoutesAll = true
					}
				}
			}
			validListenerCount++
		}
	}

	if validListenerCount == 0 {
		utils.AviLog.Errorf("key: %s, msg: Gateway %s does not contain any valid listener", key, gateway.Name)
		defaultCondition.
			Type(string(gatewayv1.GatewayConditionAccepted)).
			Reason(string(gatewayv1.GatewayReasonListenersNotValid)).
			Message("Gateway does not contain any valid listener").
			SetIn(&gatewayStatus.Conditions)
		programmedCondition.
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &status.Status{GatewayStatus: gatewayStatus})
		return false, allowedRoutesAll
	} else if validListenerCount < len(spec.Listeners) {
		defaultCondition.
			Reason(string(gatewayv1.GatewayReasonListenersNotValid)).
			Status(metav1.ConditionTrue).
			Message("Gateway contains atleast one valid listener").
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &status.Status{GatewayStatus: gatewayStatus})
		utils.AviLog.Infof("key: %s, msg: Gateway %s contains atleast one valid listener", key, gateway.Name)
		return true, allowedRoutesAll
	}

	defaultCondition.
		Reason(string(gatewayv1.GatewayReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Gateway configuration is valid").
		SetIn(&gatewayStatus.Conditions)
	akogatewayapistatus.Record(key, gateway, &status.Status{GatewayStatus: gatewayStatus})
	utils.AviLog.Infof("key: %s, msg: Gateway %s is valid", key, gateway.Name)
	return true, allowedRoutesAll
}

func isValidListener(key string, gateway *gatewayv1.Gateway, gatewayStatus *gatewayv1.GatewayStatus, index int) bool {

	listener := gateway.Spec.Listeners[index]
	gatewayStatus.Listeners[index].Name = gateway.Spec.Listeners[index].Name
	gatewayStatus.Listeners[index].SupportedKinds = akogatewayapilib.SupportedKinds[listener.Protocol]
	gatewayStatus.Listeners[index].AttachedRoutes = akogatewayapilib.ZeroAttachedRoutes

	defaultCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.ListenerConditionAccepted)).
		Reason(string(gatewayv1.ListenerReasonInvalid)).
		Message("Listener is Invalid").
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)

	programmedCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.ListenerConditionProgrammed)).
		Reason(string(gatewayv1.ListenerReasonInvalid)).
		Message("Virtual service not configured/updated for this listener").
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)

	// protocol validation
	if listener.Protocol != gatewayv1.HTTPProtocolType &&
		listener.Protocol != gatewayv1.HTTPSProtocolType {
		utils.AviLog.Errorf("key: %s, msg: protocol is not supported for listener %s", key, listener.Name)
		defaultCondition.
			Reason(string(gatewayv1.ListenerReasonUnsupportedProtocol)).
			Message("Unsupported protocol").
			SetIn(&gatewayStatus.Listeners[index].Conditions)
		programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
		gatewayStatus.Listeners[index].SupportedKinds = akogatewayapilib.SupportedKinds[gatewayv1.HTTPSProtocolType]
		return false
	}

	// hostname should not overlap with hostname of an existing gateway
	gatewayNsList, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(gateway.Namespace).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the gateways during validation: %s", err)
		return false
	}
	for _, gatewayInNamespace := range gatewayNsList {
		if gateway.Name != gatewayInNamespace.Name {
			for _, gwListener := range gatewayInNamespace.Spec.Listeners {
				if gwListener.Hostname == nil {
					continue
				}
				if listener.Hostname != nil && *listener.Hostname == *gwListener.Hostname {
					utils.AviLog.Errorf("key: %s, msg: Hostname is same as an existing gateway %s hostname %s", key, gatewayInNamespace.Name, *gwListener.Hostname)
					defaultCondition.
						Message("Hostname is same as an existing gateway hostname").
						SetIn(&gatewayStatus.Listeners[index].Conditions)
					programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
					return false
				}
			}
		}
	}
	// do not check subdomain for empty or * hostname
	if listener.Hostname != nil && *listener.Hostname != utils.WILDCARD && *listener.Hostname != "" {
		if !akogatewayapilib.VerifyHostnameSubdomainMatch(string(*listener.Hostname)) {
			defaultCondition.
				Message(fmt.Sprintf("Didn't find match for hostname :%s in available sub-domains", string(*listener.Hostname))).
				SetIn(&gatewayStatus.Listeners[index].Conditions)
			programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
			return false
		}
	}

	resolvedRefCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.ListenerConditionResolvedRefs)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)
	// has valid TLS config
	if listener.TLS != nil {
		if (listener.TLS.Mode != nil && *listener.TLS.Mode != gatewayv1.TLSModeTerminate) || len(listener.TLS.CertificateRefs) == 0 {
			utils.AviLog.Errorf("key: %s, msg: tls mode/ref not valid %+v/%+v", key, gateway.Name, listener.Name)
			defaultCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
			resolvedRefCondition.Reason(string(gatewayv1.ListenerReasonInvalidCertificateRef)).
				Message("TLS mode or reference not valid").
				SetIn(&gatewayStatus.Listeners[index].Conditions)
			programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
			return false
		}
		for _, certRef := range listener.TLS.CertificateRefs {
			//only secret is allowed
			if (certRef.Group != nil && string(*certRef.Group) != "") ||
				certRef.Kind != nil && string(*certRef.Kind) != utils.Secret {
				utils.AviLog.Errorf("key: %s, msg: CertificateRef is not valid %+v/%+v, must be Secret", key, gateway.Name, listener.Name)
				defaultCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
				resolvedRefCondition.Reason(string(gatewayv1.ListenerReasonInvalidCertificateRef)).
					Message("TLS mode or reference not valid").
					SetIn(&gatewayStatus.Listeners[index].Conditions)
				programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
				return false
			}
			name := string(certRef.Name)
			_, err := utils.GetInformers().ClientSet.CoreV1().Secrets(gateway.ObjectMeta.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				utils.AviLog.Errorf("key: %s, msg: Secret specified in CertificateRef does not exist %+v/%+v", key, gateway.Name, listener.Name)
				gWNSName := gateway.ObjectMeta.Namespace + "/" + gateway.ObjectMeta.Name
				secretNSName := gateway.ObjectMeta.Namespace + "/" + name
				akogatewayapiobjects.GatewayApiLister().UpdateSecretToGateway(secretNSName, []string{gWNSName})
				defaultCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
				resolvedRefCondition.
					Reason(string(gatewayv1.ListenerReasonInvalidCertificateRef)).
					Message("Secret does not exist").
					SetIn(&gatewayStatus.Listeners[index].Conditions)
				programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
				return false
			}
		}
	}

	//allowedRoutes validation
	if listener.AllowedRoutes != nil {
		if listener.AllowedRoutes.Kinds != nil {
			for _, kindInAllowedRoute := range listener.AllowedRoutes.Kinds {
				if kindInAllowedRoute.Kind != "" && string(kindInAllowedRoute.Kind) != utils.HTTPRoute {
					utils.AviLog.Errorf("key: %s, msg: AllowedRoute kind is invalid %+v/%+v. Supported AllowedRoute kind is HTTPRoute.", key, gateway.Name, listener.Name)
					defaultCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
					resolvedRefCondition.
						Reason(string(gatewayv1.ListenerReasonInvalidRouteKinds)).
						Message("AllowedRoute kind is invalid. Only HTTPRoute is supported currently").
						SetIn(&gatewayStatus.Listeners[index].Conditions)
					programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
					return false
				}
				if kindInAllowedRoute.Group != nil && *kindInAllowedRoute.Group != "" && string(*kindInAllowedRoute.Group) != gatewayv1.GroupName {
					utils.AviLog.Errorf("key: %s, msg: AllowedRoute Group is invalid %+v/%+v.", key, gateway.Name, listener.Name)
					defaultCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
					resolvedRefCondition.
						Reason(string(gatewayv1.ListenerReasonInvalidRouteKinds)).
						Message("AllowedRoute Group is invalid.").
						SetIn(&gatewayStatus.Listeners[index].Conditions)
					programmedCondition.SetIn(&gatewayStatus.Listeners[index].Conditions)
					return false
				}
			}
		}
	}

	// Valid listener
	defaultCondition.
		Reason(string(gatewayv1.ListenerReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Listener is valid").
		SetIn(&gatewayStatus.Listeners[index].Conditions)

	// Setting the resolvedRef condition
	resolvedRefCondition.
		Status(metav1.ConditionTrue).
		Reason(string(gatewayv1.ListenerReasonResolvedRefs)).
		Message("All the references are valid").
		SetIn(&gatewayStatus.Listeners[index].Conditions)

	utils.AviLog.Infof("key: %s, msg: Listener %s/%s is valid", key, gateway.Name, listener.Name)
	return true
}

func IsHTTPRouteConfigValid(key string, obj *gatewayv1.HTTPRoute) bool {

	httpRoute := obj.DeepCopy()
	if len(httpRoute.Spec.ParentRefs) == 0 {
		utils.AviLog.Errorf("key: %s, msg: Parent Reference is empty for the HTTPRoute %s", key, httpRoute.Name)
		return false
	}
	return true
}

func ValidateGatewayListenerWithSecret(key, namespace, name string, deleteFlag bool) {
	secretNSName := namespace + "/" + name
	present, gwList := akogatewayapiobjects.GatewayApiLister().GetSecretToGateway(secretNSName)
	utils.AviLog.Debugf("key: %s, msg: Gateways associated with secret [%s/%s] are [%v]", key, namespace, name, gwList)
	if present {
		clonedGWList := make([]string, len(gwList))
		copy(clonedGWList, gwList)
		for _, gwNSName := range gwList {
			gwNamespace, _, gwName := lib.ExtractTypeNameNamespace(gwNSName)
			// Check gateway present or not
			gatewayObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(gwNamespace).Get(gwName)
			if err != nil {
				// ignore gateway
				continue
			}
			gwStatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gwNamespace + "/" + gwName)

			// go through listeners, check cert is present there or not.
			validListenerCount := 0
			certFound := false
			for listenerIndex, listener := range gatewayObj.Spec.Listeners {
				if listener.TLS != nil {
					if (listener.TLS.Mode != nil && *listener.TLS.Mode != gatewayv1.TLSModeTerminate) || len(listener.TLS.CertificateRefs) == 0 {
						continue
					}
					for _, certRef := range listener.TLS.CertificateRefs {
						// TODO: In future, ref grant impact needs to be considered.
						certNamespace := namespace
						if certRef.Namespace != nil {
							certNamespace = string(*certRef.Namespace)
						}
						// add condition for checking gateway status
						if gwStatus != nil && string(certRef.Name) == name && certNamespace == namespace {
							setListenerConditions(gwStatus, listenerIndex, gatewayObj.ObjectMeta.Generation, deleteFlag)
							certFound = true
							break
						}
					}
				}
				if gwStatus != nil && !akogatewayapilib.IsListenerInvalid(gwStatus, listenerIndex) {
					validListenerCount += 1
				}
			} // listener loop ends
			// if cert not found, that means it is stale entry in secret to gateway map
			if certFound {
				setGatewayCondition(gwStatus, validListenerCount, len(gatewayObj.Spec.Listeners), gatewayObj.ObjectMeta.Generation)
			} else {
				//remove the entry from the clone
				clonedGWList = utils.Remove(clonedGWList, gwNSName)
			}
		} // gw loop ends
		// update the mapping
		utils.AviLog.Debugf("key: %s, msg: after validations, Gateways associated with secret [%s/%s] are [%v]", key, namespace, name, gwList)
		akogatewayapiobjects.GatewayApiLister().UpdateSecretToGateway(secretNSName, clonedGWList)
	}
}

// Mapping has to be taken care between secret and gateway
func setListenerConditions(gwStatus *gatewayv1.GatewayStatus, index int, generation int64, isDelete bool) {
	if !isDelete {
		akogatewayapistatus.NewCondition().
			Type(string(gatewayv1.ListenerConditionAccepted)).
			Reason(string(gatewayv1.ListenerReasonAccepted)).
			Message("Listener is valid").
			Status(metav1.ConditionTrue).
			ObservedGeneration(generation).SetIn(&gwStatus.Listeners[index].Conditions)

		akogatewayapistatus.NewCondition().
			Type(string(gatewayv1.ListenerConditionResolvedRefs)).
			Reason(string(gatewayv1.ListenerReasonResolvedRefs)).
			Message("Reference is valid").
			Status(metav1.ConditionTrue).
			ObservedGeneration(generation).SetIn(&gwStatus.Listeners[index].Conditions)
		return
	}
	akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.ListenerConditionAccepted)).
		Reason(string(gatewayv1.ListenerReasonInvalid)).
		Message("Listener is Invalid").
		Status(metav1.ConditionFalse).
		ObservedGeneration(generation).SetIn(&gwStatus.Listeners[index].Conditions)

	akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.ListenerConditionResolvedRefs)).
		Reason(string(gatewayv1.ListenerReasonInvalidCertificateRef)).
		Message("Secret does not exist").
		Status(metav1.ConditionFalse).
		ObservedGeneration(generation).SetIn(&gwStatus.Listeners[index].Conditions)
}

func setGatewayCondition(gwStatus *gatewayv1.GatewayStatus, validListenerCount, totalListenerCount int, observedGeneration int64) {
	if validListenerCount == 0 {
		akogatewayapistatus.NewCondition().
			Type(string(gatewayv1.GatewayConditionAccepted)).
			Reason(string(gatewayv1.GatewayReasonListenersNotValid)).
			Message("Gateway does not contain any valid listener").
			Status(metav1.ConditionFalse).
			ObservedGeneration(observedGeneration).SetIn(&gwStatus.Conditions)
		return
	}
	if validListenerCount < totalListenerCount {
		akogatewayapistatus.NewCondition().
			Type(string(gatewayv1.GatewayConditionAccepted)).
			Reason(string(gatewayv1.GatewayReasonListenersNotValid)).
			Message("Gateway contains atleast one valid listener").
			Status(metav1.ConditionTrue).
			ObservedGeneration(observedGeneration).SetIn(&gwStatus.Conditions)
		return
	}
	akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.GatewayConditionAccepted)).
		Reason(string(gatewayv1.GatewayReasonAccepted)).
		Message("Gateway configuration is valid").
		Status(metav1.ConditionTrue).
		ObservedGeneration(observedGeneration).SetIn(&gwStatus.Conditions)
}

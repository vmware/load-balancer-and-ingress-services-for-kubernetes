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
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"k8s.io/apimachinery/pkg/labels"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapistatus "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func isRegexMatch(stringWithWildCard string, stringToBeMatched string, key string) bool {
	// replace the wildcard character with a regex
	replacedHostname := strings.Replace(stringWithWildCard, "*", "([a-zA-Z0-9-]{1,})", 1)
	// create the expression for pattern matching
	pattern := fmt.Sprintf("^%s$", replacedHostname)
	expr, err := regexp.Compile(pattern)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to compile wildcard string to regex object. Err: %s", key, err)
	}
	return expr.MatchString(stringToBeMatched)
}

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
	akogatewayapistatus.Record(key, gatewayClass, &akogatewayapistatus.Status{GatewayClassStatus: gatewayClassStatus})
	utils.AviLog.Infof("key: %s, msg: GatewayClass object %s is valid", key, gatewayClass.Name)
	return true
}

func IsValidGateway(key string, gateway *gatewayv1.Gateway) bool {
	spec := gateway.Spec

	defaultCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.GatewayConditionAccepted)).
		Reason(string(gatewayv1.GatewayReasonInvalid)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)

	gatewayStatus := gateway.Status.DeepCopy()

	// has 1 or more listeners
	if len(spec.Listeners) == 0 {
		utils.AviLog.Errorf("key: %s, msg: no listeners found in gateway %+v", key, gateway.Name)
		defaultCondition.
			Message("No listeners found").
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &akogatewayapistatus.Status{GatewayStatus: gatewayStatus})
		return false
	}

	// has 1 or none addresses
	if len(spec.Addresses) > 1 {
		utils.AviLog.Errorf("key: %s, msg: more than 1 gateway address found in gateway %+v", key, gateway.Name)
		defaultCondition.
			Message("More than one address is not supported").
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &akogatewayapistatus.Status{GatewayStatus: gatewayStatus})
		return false
	}

	if len(spec.Addresses) == 1 && *spec.Addresses[0].Type != "IPAddress" {
		utils.AviLog.Errorf("key: %s, msg: gateway address is not of type IPAddress %+v", key, gateway.Name)
		defaultCondition.
			Message("Only IPAddress as AddressType is supported").
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &akogatewayapistatus.Status{GatewayStatus: gatewayStatus})
		return false
	}

	gatewayStatus.Listeners = make([]gatewayv1.ListenerStatus, len(gateway.Spec.Listeners))

	var invalidListenerCount int
	for index := range spec.Listeners {
		if !isValidListener(key, gateway, gatewayStatus, index) {
			invalidListenerCount++
		}
	}

	if invalidListenerCount > 0 {
		utils.AviLog.Errorf("key: %s, msg: Gateway %s contains %d invalid listeners", key, gateway.Name, invalidListenerCount)
		defaultCondition.
			Type(string(gatewayv1.GatewayConditionAccepted)).
			Reason(string(gatewayv1.GatewayReasonListenersNotValid)).
			Message(fmt.Sprintf("Gateway contains %d invalid listener(s)", invalidListenerCount)).
			SetIn(&gatewayStatus.Conditions)
		akogatewayapistatus.Record(key, gateway, &akogatewayapistatus.Status{GatewayStatus: gatewayStatus})
		return false
	}

	defaultCondition.
		Reason(string(gatewayv1.GatewayReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Gateway configuration is valid").
		SetIn(&gatewayStatus.Conditions)
	akogatewayapistatus.Record(key, gateway, &akogatewayapistatus.Status{GatewayStatus: gatewayStatus})
	utils.AviLog.Infof("key: %s, msg: Gateway %s is valid", key, gateway.Name)
	return true
}

func isValidListener(key string, gateway *gatewayv1.Gateway, gatewayStatus *gatewayv1.GatewayStatus, index int) bool {

	listener := gateway.Spec.Listeners[index]
	gatewayStatus.Listeners[index].Name = gateway.Spec.Listeners[index].Name
	gatewayStatus.Listeners[index].SupportedKinds = akogatewayapilib.SupportedKinds[listener.Protocol]
	gatewayStatus.Listeners[index].AttachedRoutes = akogatewayapilib.ZeroAttachedRoutes

	defaultCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.GatewayConditionAccepted)).
		Reason(string(gatewayv1.GatewayReasonListenersNotValid)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(gateway.ObjectMeta.Generation)

	// hostname is not nil or wildcard
	if listener.Hostname == nil || *listener.Hostname == "*" {
		utils.AviLog.Errorf("key: %s, msg: hostname is either wildcard or empty in listener %s", key, listener.Name)
		defaultCondition.
			Message("Hostname not found or Hostname has invalid configuration").
			SetIn(&gatewayStatus.Listeners[index].Conditions)
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
				if *listener.Hostname == *gwListener.Hostname || isRegexMatch(string(*listener.Hostname), string(*gwListener.Hostname), key) || isRegexMatch(string(*gwListener.Hostname), string(*listener.Hostname), key) {
					utils.AviLog.Errorf("key: %s, msg: Hostname overlaps or is same as an existing gateway %s hostname %s", key, gatewayInNamespace.Name, *gwListener.Hostname)
					defaultCondition.
						Message("Hostname overlaps or is same as an existing gateway hostname").
						SetIn(&gatewayStatus.Listeners[index].Conditions)
					return false
				}
			}
		}
	}
	if !akogatewayapilib.VerifyHostnameSubdomainMatch(string(*listener.Hostname)) {
		defaultCondition.
			Message(fmt.Sprintf("Didn't find match for hostname :%s in available sub-domains", string(*listener.Hostname))).
			SetIn(&gatewayStatus.Listeners[index].Conditions)
		return false
	}

	// protocol validation
	if listener.Protocol != gatewayv1.HTTPProtocolType &&
		listener.Protocol != gatewayv1.HTTPSProtocolType {
		utils.AviLog.Errorf("key: %s, msg: protocol is not supported for listener %s", key, listener.Name)
		defaultCondition.
			Reason(string(gatewayv1.ListenerReasonUnsupportedProtocol)).
			Message("Unsupported protocol").
			SetIn(&gatewayStatus.Listeners[index].Conditions)
		gatewayStatus.Listeners[index].SupportedKinds = akogatewayapilib.SupportedKinds[gatewayv1.HTTPSProtocolType]
		return false
	}

	// has valid TLS config
	if listener.TLS != nil {
		if (listener.TLS.Mode != nil && *listener.TLS.Mode != gatewayv1.TLSModeTerminate) || len(listener.TLS.CertificateRefs) == 0 {
			utils.AviLog.Errorf("key: %s, msg: tls mode/ref not valid %+v/%+v", key, gateway.Name, listener.Name)
			defaultCondition.
				Reason(string(gatewayv1.ListenerReasonInvalidCertificateRef)).
				Message("TLS mode or reference not valid").
				SetIn(&gatewayStatus.Listeners[index].Conditions)
			return false
		}
		for _, certRef := range listener.TLS.CertificateRefs {
			//only secret is allowed
			if (certRef.Group != nil && string(*certRef.Group) != "") ||
				certRef.Kind != nil && string(*certRef.Kind) != utils.Secret {
				utils.AviLog.Errorf("key: %s, msg: CertificateRef is not valid %+v/%+v, must be Secret", key, gateway.Name, listener.Name)
				defaultCondition.
					Reason(string(gatewayv1.ListenerReasonInvalidCertificateRef)).
					Message("TLS mode or reference not valid").
					SetIn(&gatewayStatus.Listeners[index].Conditions)
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
					defaultCondition.
						Type(string(gatewayv1.ListenerConditionResolvedRefs)).
						Reason(string(gatewayv1.ListenerReasonInvalidRouteKinds)).
						Message("AllowedRoute kind is invalid. Only HTTPRoute is supported currently").
						SetIn(&gatewayStatus.Listeners[index].Conditions)
					return false
				}
				if kindInAllowedRoute.Group != nil && *kindInAllowedRoute.Group != "" && string(*kindInAllowedRoute.Group) != gatewayv1.GroupName {
					utils.AviLog.Errorf("key: %s, msg: AllowedRoute Group is invalid %+v/%+v.", key, gateway.Name, listener.Name)
					defaultCondition.
						Type(string(gatewayv1.ListenerConditionResolvedRefs)).
						Reason(string(gatewayv1.ListenerReasonInvalidRouteKinds)).
						Message("AllowedRoute Group is invalid.").
						SetIn(&gatewayStatus.Listeners[index].Conditions)
					return false
				}
			}
		}
	}

	// Valid listener
	defaultCondition.
		Reason(string(gatewayv1.GatewayReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Listener is valid").
		SetIn(&gatewayStatus.Listeners[index].Conditions)
	utils.AviLog.Infof("key: %s, msg: Listener %s/%s is valid", key, gateway.Name, listener.Name)
	return true
}

func IsHTTPRouteValid(key string, obj *gatewayv1.HTTPRoute) bool {

	httpRoute := obj.DeepCopy()
	if len(httpRoute.Spec.ParentRefs) == 0 {
		utils.AviLog.Errorf("key: %s, msg: Parent Reference is empty for the HTTPRoute %s", key, httpRoute.Name)
		return false
	}

	for _, hostname := range httpRoute.Spec.Hostnames {
		if strings.Contains(string(hostname), "*") {
			utils.AviLog.Errorf("key: %s, msg: Wildcard in hostname is not supported for the HTTPRoute %s", key, httpRoute.Name)
			akogatewayapilib.AKOControlConfig().EventRecorder().Eventf(httpRoute, corev1.EventTypeWarning,
				lib.Detached, "Wildcard in hostname is not supported for the HTTPRoute %s", httpRoute.Name)
			return false
		}
	}

	httpRouteStatus := obj.Status.DeepCopy()
	httpRouteStatus.Parents = make([]gatewayv1.RouteParentStatus, 0, len(httpRoute.Spec.ParentRefs))
	var invalidParentRefCount int
	isValidHttprouteRules := validateHTTPRouteRules(key, httpRoute, httpRouteStatus)
	if isValidHttprouteRules {
		for index := range httpRoute.Spec.ParentRefs {
			err := validateParentReference(key, httpRoute, httpRouteStatus, index)
			if err != nil {
				invalidParentRefCount++
				parentRefName := httpRoute.Spec.ParentRefs[index].Name
				utils.AviLog.Warnf("key: %s, msg: Parent Reference %s of HTTPRoute object %s is not valid, err: %v", key, parentRefName, httpRoute.Name, err)
			}
		}
	}
	akogatewayapistatus.Record(key, httpRoute, &akogatewayapistatus.Status{HTTPRouteStatus: httpRouteStatus})

	// No valid attachment, we can't proceed with this HTTPRoute object.
	if invalidParentRefCount == len(httpRoute.Spec.ParentRefs) || !isValidHttprouteRules {
		utils.AviLog.Errorf("key: %s, msg: HTTPRoute object %s is not valid", key, httpRoute.Name)
		akogatewayapilib.AKOControlConfig().EventRecorder().Eventf(httpRoute, corev1.EventTypeWarning,
			lib.Detached, "HTTPRoute object %s is not valid", httpRoute.Name)
		return false
	}
	utils.AviLog.Infof("key: %s, msg: HTTPRoute object %s is valid", key, httpRoute.Name)
	return true
}

func validateParentReference(key string, httpRoute *gatewayv1.HTTPRoute, httpRouteStatus *gatewayv1.HTTPRouteStatus, index int) error {

	name := string(httpRoute.Spec.ParentRefs[index].Name)
	namespace := httpRoute.Namespace
	if httpRoute.Spec.ParentRefs[index].Namespace != nil {
		namespace = string(*httpRoute.Spec.ParentRefs[index].Namespace)
	}
	gwNsName := namespace + "/" + name
	obj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: unable to get the gateway object. err: %s", key, err)
		return err
	}
	gateway := obj.DeepCopy()

	gwClass := string(gateway.Spec.GatewayClassName)
	_, isAKOCtrl := akogatewayapiobjects.GatewayApiLister().IsGatewayClassControllerAKO(gwClass)
	if !isAKOCtrl {
		utils.AviLog.Warnf("key: %s, msg: controller for the parent reference %s of HTTPRoute object %s is not ako", key, name, httpRoute.Name)
		return fmt.Errorf("controller for the parent reference %s of HTTPRoute object %s is not ako", name, httpRoute.Name)
	}
	// creates the Parent status only when the AKO is the gateway controller
	httpRouteStatus.Parents = append(httpRouteStatus.Parents, gatewayv1.RouteParentStatus{})
	httpRouteStatus.Parents[index].ControllerName = akogatewayapilib.GatewayController
	httpRouteStatus.Parents[index].ParentRef.Name = gatewayv1.ObjectName(name)
	httpRouteStatus.Parents[index].ParentRef.Namespace = (*gatewayv1.Namespace)(&namespace)
	if httpRoute.Spec.ParentRefs[index].SectionName != nil {
		httpRouteStatus.Parents[index].ParentRef.SectionName = httpRoute.Spec.ParentRefs[index].SectionName
	}

	defaultCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.GatewayConditionAccepted)).
		Reason(string(gatewayv1.GatewayReasonInvalid)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(httpRoute.ObjectMeta.Generation)

	if len(gateway.Status.Conditions) == 0 {
		// Gateway processing by AKO has not started.
		utils.AviLog.Errorf("key: %s, msg: AKO is yet to process Gateway %s for parent reference %s.", key, gateway.Name, name)
		err := fmt.Errorf("AKO is yet to process Gateway %s for parent reference %s", gateway.Name, name)
		defaultCondition.
			Message(err.Error()).
			SetIn(&httpRouteStatus.Parents[index].Conditions)
		return err
	}

	// Attach only when gateway configuration is valid
	currentGatewayStatusCondition := gateway.Status.Conditions[0]
	if currentGatewayStatusCondition.Status != metav1.ConditionTrue {
		// Gateway is not in an expected state.
		utils.AviLog.Errorf("key: %s, msg: Gateway %s for parent reference %s is in Invalid State", key, gateway.Name, name)
		err := fmt.Errorf("Gateway %s is in Invalid State", gateway.Name)
		defaultCondition.
			Message(err.Error()).
			SetIn(&httpRouteStatus.Parents[index].Conditions)
		return err
	}

	//section name is optional
	var listenersForRoute []gatewayv1.Listener
	if httpRoute.Spec.ParentRefs[index].SectionName != nil {
		listenerName := *httpRoute.Spec.ParentRefs[index].SectionName
		i := akogatewayapilib.FindListenerByName(string(listenerName), gateway.Spec.Listeners)
		if i == -1 {
			// listener is not present in gateway
			utils.AviLog.Errorf("key: %s, msg: unable to find the listener from the Section Name %s in Parent Reference %s", key, name, listenerName)
			err := fmt.Errorf("Invalid listener name provided")
			defaultCondition.
				Message(err.Error()).
				SetIn(&httpRouteStatus.Parents[index].Conditions)
			return err
		}
		listenersForRoute = append(listenersForRoute, gateway.Spec.Listeners[i])
	} else {
		listenersForRoute = append(listenersForRoute, gateway.Spec.Listeners...)
	}

	var listenersMatchedToRoute []gatewayv1.Listener
	for _, listenerObj := range listenersForRoute {
		// check from store
		hostInListener := listenerObj.Hostname

		if hostInListener == nil {
			utils.AviLog.Errorf("key: %s, msg: no hostname found in parent", key)
			continue
		}
		// replace the wildcard character with a regex
		replacedHostname := strings.Replace(string(*hostInListener), "*", "([a-zA-Z0-9-]{1,})", 1)
		// create the expression for pattern matching
		pattern := fmt.Sprintf("^%s$", replacedHostname)
		expr, err := regexp.Compile(pattern)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: unable to match the hostname with listener hostname. err: %s", key, err)
			continue
		}
		var matched bool
		for _, host := range httpRoute.Spec.Hostnames {
			matched = matched || expr.MatchString(string(host))
		}
		if !matched {
			utils.AviLog.Warnf("key: %s, msg: Gateway object %s don't have any listeners that matches the hostnames in HTTPRoute %s", key, gateway.Name, httpRoute.Name)
			continue
		}
		listenersMatchedToRoute = append(listenersMatchedToRoute, listenerObj)
	}
	if len(listenersMatchedToRoute) == 0 {
		err := fmt.Errorf("Hostname in Gateway Listener doesn't match with any of the hostnames in HTTPRoute")
		defaultCondition.
			Message(err.Error()).
			SetIn(&httpRouteStatus.Parents[index].Conditions)
		found, hosts := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(gwNsName)
		if found {
			utils.AviLog.Warnf("key: %s, msg: Hostname in Gateway Listener doesn't match with any of the hostnames in HTTPRoute", key)
			utils.AviLog.Debugf("key: %s, msg: %d hosts mapped to the route %s/%s/%s", key, len(hosts), "HTTPRoute", httpRoute.Namespace, httpRoute.Name)
			return nil
		}
		return err
	}
	gatewayStatus := gateway.Status.DeepCopy()
	for _, listenerObj := range listenersMatchedToRoute {
		listenerName := listenerObj.Name
		// Increment the attached routes of the listener in the Gateway object

		i := akogatewayapilib.FindListenerStatusByName(string(listenerName), gatewayStatus.Listeners)
		if i == -1 {
			utils.AviLog.Errorf("key: %s, msg: Gateway status is missing for the listener with name %s", key, listenerName)
			err := fmt.Errorf("Couldn't find the listener %s in the Gateway status", listenerName)
			defaultCondition.
				Message(err.Error()).
				SetIn(&httpRouteStatus.Parents[index].Conditions)
			return err
		}

		gatewayStatus.Listeners[i].AttachedRoutes += 1
	}
	akogatewayapistatus.Record(key, gateway, &akogatewayapistatus.Status{GatewayStatus: gatewayStatus})

	defaultCondition.
		Reason(string(gatewayv1.GatewayReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Parent reference is valid").
		SetIn(&httpRouteStatus.Parents[index].Conditions)
	utils.AviLog.Infof("key: %s, msg: Parent Reference %s of HTTPRoute object %s is valid", key, name, httpRoute.Name)
	return nil
}
func validateHTTPRouteRules(key string, httpRoute *gatewayv1.HTTPRoute, httpRouteStatus *gatewayv1.HTTPRouteStatus) bool {

	//Validate Filters

	//Validate URL Rewrite Filter
	if httpRoute.Spec.Rules != nil {
		for _, rule := range httpRoute.Spec.Rules {
			for _, filter := range rule.Filters {
				if filter.Type == gatewayv1.HTTPRouteFilterURLRewrite && filter.URLRewrite != nil && filter.URLRewrite.Path != nil && filter.URLRewrite.Path.Type != gatewayv1.FullPathHTTPPathModifier {
					for parentRefIndexFromSpec := range httpRoute.Spec.ParentRefs {

						// creates the Parent status only when the AKO is the gateway controller
						name := string(httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Name)
						namespace := httpRoute.Namespace
						if httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Namespace != nil {
							namespace = string(*httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Namespace)
						}
						obj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
						if err != nil {
							utils.AviLog.Errorf("key: %s, msg: unable to get the gateway object. err: %s", key, err)
							return false
						}
						gateway := obj.DeepCopy()

						gwClass := string(gateway.Spec.GatewayClassName)
						_, isAKOCtrl := akogatewayapiobjects.GatewayApiLister().IsGatewayClassControllerAKO(gwClass)
						if !isAKOCtrl {
							utils.AviLog.Warnf("key: %s, msg: controller for the parent reference %s of HTTPRoute object %s is not ako", key, name, httpRoute.Name)
							continue
						}

						httpRouteStatus.Parents = append(httpRouteStatus.Parents, gatewayv1.RouteParentStatus{})
						httpRouteStatus.Parents[parentRefIndexFromSpec].ControllerName = akogatewayapilib.GatewayController
						httpRouteStatus.Parents[parentRefIndexFromSpec].ParentRef.Name = gatewayv1.ObjectName(name)
						httpRouteStatus.Parents[parentRefIndexFromSpec].ParentRef.Namespace = (*gatewayv1.Namespace)(&namespace)
						if httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].SectionName != nil {
							httpRouteStatus.Parents[parentRefIndexFromSpec].ParentRef.SectionName = httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].SectionName
						}
						routeConditionAccepted := akogatewayapistatus.NewCondition().
							Type(string(gatewayv1.RouteConditionAccepted)).
							Status(metav1.ConditionFalse).
							ObservedGeneration(httpRoute.ObjectMeta.Generation).
							Reason(string(gatewayv1.RouteReasonUnsupportedValue)).
							Message("HTTPUrlRewrite PathType has Unsupported value")
						routeConditionAccepted.SetIn(&httpRouteStatus.Parents[parentRefIndexFromSpec].Conditions)
					}
					utils.AviLog.Errorf("key: %s, msg: HTTPUrlRewrite PathType has Unsupported value %s.", key, filter.URLRewrite.Path.Type)
					return false

				}
			}
		}
	}

	return true
}

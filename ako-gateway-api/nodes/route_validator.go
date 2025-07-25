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

package nodes

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapistatus "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var SupportedExtensionRefKindsOnHTTPRouteRule = map[string]string{
	"L7Rule":             "L7Rule",
	"ApplicationProfile": "ApplicationProfile",
}
var SupportedExtensionRefKindsOnHTTPRouteBackendRef = map[string]string{
	"RouteBackendExtension": "RouteBackendExtension",
	"HealthMonitor":         "HealthMonitor",
}

func isRegexMatch(stringWithWildCard string, stringToBeMatched string, key string) bool {
	// replace the wildcard character with a regex
	replacedHostname := strings.Replace(stringWithWildCard, utils.WILDCARD, utils.FQDN_LABEL_REGEX, 1)
	// create the expression for pattern matching
	pattern := fmt.Sprintf("^%s$", replacedHostname)
	expr, err := regexp.Compile(pattern)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to compile wildcard string to regex object. Err: %s", key, err)
	}
	return expr.MatchString(stringToBeMatched)
}

func IsHTTPRouteValid(key string, obj *gatewayv1.HTTPRoute) bool {

	httpRoute := obj.DeepCopy()
	httpRouteStatus := obj.Status.DeepCopy()
	httpRouteStatus.Parents = make([]gatewayv1.RouteParentStatus, 0, len(httpRoute.Spec.ParentRefs))
	var invalidParentRefCount int
	parentRefIndexInHttpRouteStatus := 0
	indexInCache := -1
	isValidHttprouteRules := validateHTTPRouteRules(key, httpRoute, httpRouteStatus)
	if isValidHttprouteRules {
		for parentRefIndexFromSpec := range httpRoute.Spec.ParentRefs {
			err := validateParentReference(key, httpRoute, httpRouteStatus, parentRefIndexFromSpec, &parentRefIndexInHttpRouteStatus, &indexInCache)
			if err != nil {
				invalidParentRefCount++
				parentRefName := httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Name
				utils.AviLog.Warnf("key: %s, msg: Parent Reference %s of HTTPRoute object %s is not valid, err: %v", key, parentRefName, httpRoute.Name, err)
			}
		}
	}

	akogatewayapistatus.Record(key, httpRoute, &status.Status{HTTPRouteStatus: httpRouteStatus})

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

func validateBackendReference(key string, backend Backend, backendFilters []*Filter, httpRouteNamespace string) (bool, akogatewayapistatus.Condition) {
	routeConditionResolvedRef := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.RouteConditionResolvedRefs)).
		Status(metav1.ConditionFalse)
	if backend.Kind != "" && backend.Kind != "Service" {
		utils.AviLog.Errorf("key: %s, msg: BackendRef %s has invalid kind %s.", key, backend.Name, backend.Kind)
		err := fmt.Errorf("backendRef %s has invalid kind %s", backend.Name, backend.Kind)
		routeConditionResolvedRef.
			Reason(string(gatewayv1.RouteReasonInvalidKind)).
			Message(err.Error())
		return false, routeConditionResolvedRef
	}
	backendRefTenant := lib.GetTenantInNamespace(backend.Namespace)
	httpRouteTenant := lib.GetTenantInNamespace(httpRouteNamespace)

	// check if tenant of namespace of backend is same as tenant of namespace of HTTPRoute
	if backendRefTenant != httpRouteTenant {
		utils.AviLog.Errorf("key: %s, msg: BackendRef %s tenant %s is not equal to HTTPRoute tenant %s", key, backend.Name, backendRefTenant, httpRouteTenant)
		err := fmt.Errorf("backendRef %s tenant %s is not equal to HTTPRoute tenant %s", backend.Name, backendRefTenant, httpRouteTenant)
		// using RouteReasonRefNotPermitted for now, may change when reference grant is supported
		routeConditionResolvedRef.
			Reason(string(gatewayv1.RouteReasonRefNotPermitted)).
			Message(err.Error())
		return false, routeConditionResolvedRef
	}

	flag, routeConditionResolvedRef := validatedBackendRefExtensions(backendFilters, routeConditionResolvedRef, key, backend)
	if !flag {
		return false, routeConditionResolvedRef
	}
	// Valid route case
	routeConditionResolvedRef.
		Status(metav1.ConditionTrue).
		Reason(string(gatewayv1.RouteReasonResolvedRefs))
	return true, routeConditionResolvedRef
}

func validatedBackendRefExtensions(backendFilters []*Filter, routeConditionResolvedRef akogatewayapistatus.Condition, key string, backend Backend) (bool, akogatewayapistatus.Condition) {
	extensionRefType := make(map[string]struct{})
	for _, filter := range backendFilters {
		// Current support only ExtensionRef
		if filter == nil {
			continue
		}
		if filter.Type != string(gatewayv1.HTTPRouteFilterExtensionRef) {
			utils.AviLog.Errorf("key: %s, msg: BackendRef %s has unsupported Filter Type %s", key, backend.Name, filter.Type)
			err := fmt.Errorf("backendRef %s has unsupported Filter Type %s", backend.Name, filter.Type)
			routeConditionResolvedRef.
				Reason(string(gatewayv1.RouteReasonUnsupportedValue)).
				Message(err.Error())
			return false, routeConditionResolvedRef
		}
		if filter.ExtensionRef != nil {
			if string(filter.ExtensionRef.Group) != lib.AkoGroup {
				utils.AviLog.Warnf("key: %s, msg: Extension Ref is not handled by AKO. Group of extension filter %s != %s ", key, filter.ExtensionRef.Group, lib.AkoGroup)
				continue
			}
			// Allow only one instance of each kind.
			// If user wants to define multiple instances of that kind, use an instnace of AKO defined CRD
			kind := string(filter.ExtensionRef.Kind)
			if _, ok := SupportedExtensionRefKindsOnHTTPRouteBackendRef[kind]; !ok {
				utils.AviLog.Warnf("key: %s, msg: AKO does not support a kind: %s on HTTPRoute-Rule-BackendRef", key, kind)
				routeConditionResolvedRef.
					Reason(string(gatewayv1.RouteReasonUnsupportedValue)).
					Message(fmt.Sprintf("Unsupported kind %s defined on HTTPRoute-Rule-BackendRef", kind))
				return false, routeConditionResolvedRef
			}
			if _, ok := extensionRefType[kind]; ok {
				// support multiple entries only if kind is HealthMonitor
				if kind != akogatewayapilib.HealthMonitorKind {
					utils.AviLog.Warnf("key: %s, msg: multiple entries for a kind %s. AKO handles only one object of each kind in ExtensionRef", key, kind)
					routeConditionResolvedRef.
						Reason(string(gatewayv1.RouteReasonIncompatibleFilters)).
						Message("MultipleExtensionRef of same kind defined on HTTPRoute-Rule-BackendRef")
					return false, routeConditionResolvedRef
				}
			}
			extensionRefType[kind] = struct{}{}
			if kind == akogatewayapilib.HealthMonitorKind {
				if filter.ExtensionRef.Name != "" {
					_, ready, err := akogatewayapilib.IsHealthMonitorProcessed(key, backend.Namespace, string(filter.ExtensionRef.Name))
					if err != nil || !ready {
						var errMsg string
						if err != nil {
							errMsg = err.Error()
						} else {
							errMsg = "HealthMonitor is not ready"
						}
						utils.AviLog.Warnf("key: %s, msg: error: HealthMonitor %s/%s will not be processed by gateway-container. err: %s", key, backend.Namespace, string(filter.ExtensionRef.Name), errMsg)
						routeConditionResolvedRef.
							Reason(string(gatewayv1.RouteReasonBackendNotFound)).
							Message(errMsg)
						return false, routeConditionResolvedRef
					}
				} else {
					utils.AviLog.Warnf("key: %s, msg: HealthMonitor ExtensionRef has empty name", key)
					routeConditionResolvedRef.
						Reason(string(gatewayv1.RouteReasonBackendNotFound)).
						Message("HealthMonitor ExtensionRef has empty name")
					return false, routeConditionResolvedRef
				}
			} else if kind == akogatewayapilib.RouteBackendExtensionKind {
				if filter.ExtensionRef.Name != "" {
					_, status, err := akogatewayapilib.IsRouteBackendExtensionProcessed(key, backend.Namespace, filter.ExtensionRef.Name)
					if err != nil {
						utils.AviLog.Warnf("key: %s, msg: RouteBackendExtension object %s/%s will not be processed by gateway-container, status: %s, err: %+v", key, backend.Namespace, filter.ExtensionRef.Name, status, err)
						routeConditionResolvedRef.
							Reason(string(gatewayv1.RouteReasonBackendNotFound)).
							Message(err.Error())
						return false, routeConditionResolvedRef
					} else if status != "Accepted" {
						utils.AviLog.Warnf("key: %s, msg: RouteBackendExtension object %s/%s will not be processed by gateway-container, status: %s", key, backend.Namespace, filter.ExtensionRef.Name, status)
						routeConditionResolvedRef.
							Reason(string(gatewayv1.RouteReasonBackendNotFound)).
							Message(fmt.Sprintf("RouteBackendExtension object %s/%s is not in Accepted state", backend.Namespace, filter.ExtensionRef.Name))
						return false, routeConditionResolvedRef
					}
				} else {
					utils.AviLog.Warnf("key: %s, msg: RouteBackendExtension ExtensionRef has empty name", key)
					routeConditionResolvedRef.
						Reason(string(gatewayv1.RouteReasonBackendNotFound)).
						Message("RouteBackendExtension ExtensionRef has empty name")
					return false, routeConditionResolvedRef
				}
			}
		}
	}
	return true, routeConditionResolvedRef
}

// Current behaviour: Any invalid Rule, that HTTPRoute object is not processed.
// TODO: Need to modify this behaviour by keeping map of rule to valid/invalid for a given route.
func validateHTTPRouteRules(key string, httpRoute *gatewayv1.HTTPRoute, httpRouteStatus *gatewayv1.HTTPRouteStatus) bool {
	//Validate Filters
	//Validate URL Rewrite Filter, ExtensionRef
	if httpRoute.Spec.Rules != nil {
		for _, rule := range httpRoute.Spec.Rules {
			extensionRefType := make(map[string]struct{})
			for _, filter := range rule.Filters {
				if filter.Type == gatewayv1.HTTPRouteFilterURLRewrite && filter.URLRewrite != nil && filter.URLRewrite.Path != nil && filter.URLRewrite.Path.Type != gatewayv1.FullPathHTTPPathModifier {
					setRouteConditionInHTTPRouteStatus(key,
						string(gatewayv1.RouteReasonUnsupportedValue),
						"HTTPUrlRewrite PathType has Unsupported value",
						httpRoute, httpRouteStatus)
					utils.AviLog.Errorf("key: %s, msg: HTTPUrlRewrite PathType has Unsupported value %s.", key, filter.URLRewrite.Path.Type)
					return false
				} else if filter.Type == gatewayv1.HTTPRouteFilterExtensionRef && filter.ExtensionRef != nil {
					// can convert to function
					// Allows only ako.vmware.com
					if string(filter.ExtensionRef.Group) != lib.AkoGroup {
						utils.AviLog.Warnf("key: %s, msg: Extension Ref is not handled by AKO. Group of extension filter %s != %s ", key, filter.ExtensionRef.Group, lib.AkoGroup)
						continue
					}
					// Allow only one instance of each kind.
					// If user wants to define multiple instances of that kind, use an instnace of AKO defined CRD
					kind := string(filter.ExtensionRef.Kind)
					if _, ok := SupportedExtensionRefKindsOnHTTPRouteRule[kind]; !ok {
						utils.AviLog.Warnf("key: %s, msg: AKO does not support a kind: %s on HTTPRoute-Rule", key, kind)
						// set the status
						setRouteConditionInHTTPRouteStatus(key,
							string(gatewayv1.RouteReasonUnsupportedValue),
							fmt.Sprintf("Unsupported kind %s defined on HTTPRoute-Rule", kind),
							httpRoute, httpRouteStatus)
						return false
					}
					if _, ok := extensionRefType[kind]; ok {
						utils.AviLog.Warnf("key: %s, msg: multiple entries for a kind %s. AKO handles only one object of each kind in ExtensionRef", key, kind)
						// set the status
						setRouteConditionInHTTPRouteStatus(key,
							string(gatewayv1.RouteReasonIncompatibleFilters),
							"MultipleExtensionRef of same kind defined on HTTPRoute-Rule",
							httpRoute, httpRouteStatus)
						return false
					}
					extensionRefType[kind] = struct{}{}
				}
			}
			if rule.SessionPersistence != nil {
				if rule.SessionPersistence.Type != nil && *rule.SessionPersistence.Type == gatewayv1.HeaderBasedSessionPersistence {
					utils.AviLog.Errorf("key: %s, msg: Header based session persistence type is not supported ", key)
					return false
				}
				if rule.SessionPersistence.SessionName == nil || *rule.SessionPersistence.SessionName == "" {
					utils.AviLog.Errorf("key: %s, msg: Session Name is needed in SessionPersistence", key)
					return false
				}
			}
		}
	}
	return true
}

func setRouteConditionInHTTPRouteStatus(key, reason, msg string, httpRoute *gatewayv1.HTTPRoute, httpRouteStatus *gatewayv1.HTTPRouteStatus) {
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
			return
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
			Reason(reason).
			Message(msg)
		routeConditionAccepted.SetIn(&httpRouteStatus.Parents[parentRefIndexFromSpec].Conditions)
	}
}

func setResolvedRefConditionInHTTPRouteStatus(key string, routeConditionResolvedRef akogatewayapistatus.Condition, routeTypeNamespaceName string) {
	if routeConditionResolvedRef == nil {
		return
	}
	httpRouteStatus := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(routeTypeNamespaceName)
	routeConditionResolvedRef.ObservedGeneration(httpRouteStatus.Parents[0].Conditions[0].ObservedGeneration)
	for parentRefIndex := range httpRouteStatus.Parents {
		routeConditionResolvedRef.SetIn(&httpRouteStatus.Parents[parentRefIndex].Conditions)
	}
	_, namespace, name := lib.ExtractTypeNameNamespace(routeTypeNamespaceName)
	httpRoute, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to extract the HTTPRoute object %s for BackendRef validation", key, name)
		return
	}
	akogatewayapistatus.Record(key, httpRoute, &status.Status{HTTPRouteStatus: httpRouteStatus})
}

func validateParentReference(key string, httpRoute *gatewayv1.HTTPRoute, httpRouteStatus *gatewayv1.HTTPRouteStatus, parentRefIndexFromSpec int, parentRefIndexInHttpRouteStatus *int, indexInCache *int) error {

	name := string(httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Name)
	namespace := httpRoute.Namespace
	if httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Namespace != nil {
		namespace = string(*httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Namespace)
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
	httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].ControllerName = akogatewayapilib.GatewayController
	httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].ParentRef.Name = gatewayv1.ObjectName(name)
	httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].ParentRef.Namespace = (*gatewayv1.Namespace)(&namespace)
	if httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].SectionName != nil {
		httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].ParentRef.SectionName = httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].SectionName
	}
	routeTypeNsName := lib.HTTPRoute + "/" + httpRoute.Namespace + "/" + httpRoute.Name
	routeStatusInCache := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(routeTypeNsName)
	if *indexInCache != -1 && routeStatusInCache != nil {
		if *indexInCache < len(routeStatusInCache.Parents) {
			if reflect.DeepEqual(routeStatusInCache.Parents[*indexInCache], httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus]) {
				*indexInCache = *indexInCache + 1
			}
		}
	}
	defaultCondition := akogatewayapistatus.NewCondition().
		Type(string(gatewayv1.RouteConditionAccepted)).
		Status(metav1.ConditionFalse).
		ObservedGeneration(httpRoute.ObjectMeta.Generation)

	gwStatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gwNsName)
	if len(gwStatus.Conditions) == 0 {
		// Gateway processing by AKO has not started.
		utils.AviLog.Errorf("key: %s, msg: AKO is yet to process Gateway %s for parent reference %s.", key, gateway.Name, name)
		err := fmt.Errorf("AKO is yet to process Gateway %s for parent reference %s", gateway.Name, name)
		defaultCondition.
			Reason(string(gatewayv1.RouteReasonPending)).
			Message(err.Error()).
			SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
		*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
		return err
	}

	// If Gateway and HTTPRoute are in different namespace, validate that both namespaces are scoped to the same tenant
	if httpRoute.Namespace != namespace {
		if lib.GetTenantInNamespace(httpRoute.Namespace) != lib.GetTenantInNamespace(namespace) {
			utils.AviLog.Errorf("key: %s, msg: Tenant mismatch between HTTPRoute %s and Parent Reference %s", key, httpRoute.GetName(), name)
			err := fmt.Errorf("Tenant mismatch between HTTPRoute %s and Parent Reference %s", httpRoute.GetName(), name)
			defaultCondition.
				Reason(string(gatewayv1.RouteReasonPending)).
				Message(err.Error()).
				SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
			*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
			return err
		}
	}

	// Attach only when gateway configuration is valid
	currentGatewayStatusCondition := gwStatus.Conditions[0]
	if currentGatewayStatusCondition.Status != metav1.ConditionTrue {
		// Gateway is not in an expected state.
		utils.AviLog.Errorf("key: %s, msg: Gateway %s for parent reference %s is in Invalid State", key, gateway.Name, name)
		err := fmt.Errorf("Gateway %s is in Invalid State", gateway.Name)
		defaultCondition.
			Reason(string(gatewayv1.RouteReasonPending)).
			Message(err.Error()).
			SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
		*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
		return err
	}

	//section name is optional
	var listenersForRoute []gatewayv1.Listener
	if httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].SectionName != nil {
		listenerName := *httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].SectionName
		i := akogatewayapilib.FindListenerByName(string(listenerName), gateway.Spec.Listeners)
		if i == -1 {
			// listener is not present in gateway
			utils.AviLog.Errorf("key: %s, msg: unable to find the listener from the Section Name %s in Parent Reference %s", key, name, listenerName)
			err := fmt.Errorf("Invalid listener name provided")
			defaultCondition.
				Reason(string(gatewayv1.RouteReasonNoMatchingParent)).
				Message(err.Error()).
				SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
			*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
			return err
		}
		if akogatewayapilib.IsListenerInvalid(gwStatus, i) {
			// listener is present in gateway but is in invalid state
			utils.AviLog.Errorf("key: %s, msg: Matching gateway listener %s in Parent Reference is in invalid state", key, listenerName)
			err := fmt.Errorf("Matching gateway listener is in Invalid state")
			defaultCondition.
				Reason(string(gatewayv1.RouteReasonPending)).
				Message(err.Error()).
				SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
			*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
			return err
		}
		listenersForRoute = append(listenersForRoute, gateway.Spec.Listeners[i])
	} else {
		listenersForRoute = append(listenersForRoute, gateway.Spec.Listeners...)
	}

	// TODO: Validation for hostname (those being fqdns) need to validate as per the K8 gateway req.
	// Here I need to check
	var listenersMatchedToRoute []gatewayv1.Listener
	for _, listenerObj := range listenersForRoute {
		// check from store
		hostInListener := listenerObj.Hostname
		isListenerFqdnWildcard := false
		matched := false
		// TODO:
		// Use case to handle for validations of hostname:
		// USe case 1: Shouldn't contain mor than 1 *
		// USe case 2: * should be at the beginning only
		if (hostInListener == nil || *hostInListener == "" || *hostInListener == utils.WILDCARD) && len(httpRoute.Spec.Hostnames) != 0 {
			matched = true
		} else {
			// mark listener fqdn if it has *
			if hostInListener != nil && strings.HasPrefix(string(*hostInListener), utils.WILDCARD) {
				isListenerFqdnWildcard = true
			}
			for _, host := range httpRoute.Spec.Hostnames {
				// casese to consider:
				// Case 1: hostname of gateway is wildcard(empty) and hostname from httproute is not wild card
				// Case 2: hostname of gateway is not wild card and hostname from httproute is wildcard
				// case 3: hostname of gateway is wildcard(empty) and hostname from httproute is wildcard
				// case 4: hostname of gateway is not wildcard and hostname from httproute is not wildcard
				isHttpRouteHostFqdnWildcard := false
				if strings.HasPrefix(string(host), utils.WILDCARD) {
					isHttpRouteHostFqdnWildcard = true
				}
				if isHttpRouteHostFqdnWildcard && isListenerFqdnWildcard {
					// both are true. Match nonwildcard part
					// Use case: 1. GW: *.avi.internal HttpRoute: *.bar.avi.internal
					// USe case: 2. GW: *.bar.avi.internal HttpRoute: *.avi.internal
					if utils.CheckSubdomainOverlapping(string(host), string(*hostInListener)) {
						matched = true
						break
					}

				} else if !isHttpRouteHostFqdnWildcard && !isListenerFqdnWildcard {
					// both are complete fqdn
					if string(host) == string(*hostInListener) {
						matched = true
						break
					}
				} else {
					if isHttpRouteHostFqdnWildcard {
						// httpRoute hostFqdn is wildcard
						matched = matched || isRegexMatch(string(host), string(*hostInListener), key)
					} else if isListenerFqdnWildcard {
						// listener fqdn is wildcard
						matched = matched || isRegexMatch(string(*hostInListener), string(host), key)
					}

				}

			}
			// if there are no hostnames specified, all parent listneres should be matched.
			if len(httpRoute.Spec.Hostnames) == 0 && (hostInListener != nil && *hostInListener != "" && *hostInListener != utils.WILDCARD) {
				matched = true
			}
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
			Reason(string(gatewayv1.RouteReasonNoMatchingListenerHostname)).
			Message(err.Error()).
			SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
		*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
		gwRouteNsName := fmt.Sprintf("%s/%s/%s/%s", gwNsName, lib.HTTPRoute, httpRoute.Namespace, httpRoute.Name)
		found, hosts := akogatewayapiobjects.GatewayApiLister().GetGatewayRouteToHostname(gwRouteNsName)
		if found {
			utils.AviLog.Warnf("key: %s, msg: Hostname in Gateway Listener doesn't match with any of the hostnames in HTTPRoute", key)
			utils.AviLog.Debugf("key: %s, msg: %d hosts mapped to the route %s/%s/%s", key, len(hosts), "HTTPRoute", httpRoute.Namespace, httpRoute.Name)
			return nil
		}
		return err
	}

	//TODO: Add a condition to check whether this route is allowed by the parent gateways allowedroute field and set gatewayv1.RouteReasonNotAllowedByListeners reason while implemenating gateway->listener->allowedRoutes->Selector

	gatewayStatus := gwStatus.DeepCopy()
	for _, listenerObj := range listenersMatchedToRoute {
		listenerName := listenerObj.Name
		// Increment the attached routes of the listener in the Gateway object

		i := akogatewayapilib.FindListenerStatusByName(string(listenerName), gatewayStatus.Listeners)
		if i == -1 {
			utils.AviLog.Errorf("key: %s, msg: Gateway status is missing for the listener with name %s", key, listenerName)
			err := fmt.Errorf("Couldn't find the listener %s in the Gateway status", listenerName)
			defaultCondition.
				Reason(string(gatewayv1.RouteReasonNoMatchingParent)).
				Message(err.Error()).
				SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
			*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
			return err
		}

		gatewayStatus.Listeners[i].AttachedRoutes += 1
	}
	akogatewayapistatus.Record(key, gateway, &status.Status{GatewayStatus: gatewayStatus})

	defaultCondition.
		Reason(string(gatewayv1.RouteReasonAccepted)).
		Status(metav1.ConditionTrue).
		Message("Parent reference is valid").
		SetIn(&httpRouteStatus.Parents[*parentRefIndexInHttpRouteStatus].Conditions)
	utils.AviLog.Infof("key: %s, msg: Parent Reference %s of HTTPRoute object %s is valid", key, name, httpRoute.Name)
	*parentRefIndexInHttpRouteStatus = *parentRefIndexInHttpRouteStatus + 1
	return nil
}

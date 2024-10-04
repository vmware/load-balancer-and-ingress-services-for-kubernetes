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

package nodes

import (
	"fmt"
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
	for parentRefIndexFromSpec := range httpRoute.Spec.ParentRefs {
		err := validateParentReference(key, httpRoute, httpRouteStatus, parentRefIndexFromSpec, &parentRefIndexInHttpRouteStatus)
		if err != nil {
			invalidParentRefCount++
			parentRefName := httpRoute.Spec.ParentRefs[parentRefIndexFromSpec].Name
			utils.AviLog.Warnf("key: %s, msg: Parent Reference %s of HTTPRoute object %s is not valid, err: %v", key, parentRefName, httpRoute.Name, err)
		}
	}
	akogatewayapistatus.Record(key, httpRoute, &status.Status{HTTPRouteStatus: httpRouteStatus})

	// No valid attachment, we can't proceed with this HTTPRoute object.
	if invalidParentRefCount == len(httpRoute.Spec.ParentRefs) {
		utils.AviLog.Errorf("key: %s, msg: HTTPRoute object %s is not valid", key, httpRoute.Name)
		akogatewayapilib.AKOControlConfig().EventRecorder().Eventf(httpRoute, corev1.EventTypeWarning,
			lib.Detached, "HTTPRoute object %s is not valid", httpRoute.Name)
		return false
	}
	utils.AviLog.Infof("key: %s, msg: HTTPRoute object %s is valid", key, httpRoute.Name)
	return true
}

func validateParentReference(key string, httpRoute *gatewayv1.HTTPRoute, httpRouteStatus *gatewayv1.HTTPRouteStatus, parentRefIndexFromSpec int, parentRefIndexInHttpRouteStatus *int) error {

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
		if IsListenerInvalid(gwStatus, i) {
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
		if hostInListener == nil || *hostInListener == "" || *hostInListener == utils.WILDCARD {
			matched = true
		} else {
			// mark listener fqdn if it has *
			if strings.HasPrefix(string(*hostInListener), utils.WILDCARD) {
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
			if len(httpRoute.Spec.Hostnames) == 0 {
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

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
	"encoding/json"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	akogatewayapistatus "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func ConfigDescriptor() GraphDescriptor {
	return SupportedGraphTypes
}

type GraphDescriptor []GraphSchema

type GraphSchema struct {
	Type        string
	GetGateways func(string, string, string) ([]string, bool)
	GetRoutes   func(string, string, string) ([]string, bool)
}

func (descriptor GraphDescriptor) GetByType(name string) (GraphSchema, bool) {
	for _, schema := range descriptor {
		if schema.Type == name {
			return schema, true
		}
	}
	return GraphSchema{}, false
}

var (
	Gateway = GraphSchema{
		Type:        "Gateway",
		GetGateways: GatewayGetGw,
		GetRoutes:   GatewayToRoutes,
	}
	GatewayClass = GraphSchema{
		Type:        "GatewayClass",
		GetGateways: GatewayClassGetGw,
		GetRoutes:   NoOperation,
	}
	Secret = GraphSchema{
		Type:        "Secret",
		GetGateways: SecretToGateways,
		GetRoutes:   NoOperation,
	}
	Service = GraphSchema{
		Type:        "Service",
		GetGateways: ServiceToGateways,
		GetRoutes:   ServiceToRoutes,
	}
	EndpointSlices = GraphSchema{
		Type:        utils.Endpointslices,
		GetGateways: EndpointSlicesToGateways,
		GetRoutes:   EndpointSlicesToRoutes,
	}
	HTTPRoute = GraphSchema{
		Type:        lib.HTTPRoute,
		GetGateways: HTTPRouteToGateway,
		GetRoutes:   HTTPRouteChanges,
	}
	Pod = GraphSchema{
		Type:        "Pod",
		GetGateways: PodToGateway,
		GetRoutes:   PodToHTTPRoute,
	}
	SupportedGraphTypes = GraphDescriptor{
		Gateway,
		GatewayClass,
		Secret,
		Service,
		EndpointSlices,
		HTTPRoute,
		Pod,
	}
)

func GatewayGetGw(namespace, name, key string) ([]string, bool) {
	gwNsName := namespace + "/" + name

	gwObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		// gateway must be deleted, so remove mapping
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayFromStore(gwNsName)

		return []string{gwNsName}, true
	}

	gwClassName := string(gwObj.Spec.GatewayClassName)

	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToGatewayClass(gwNsName, gwClassName)

	var listeners []akogatewayapiobjects.GatewayListenerStore
	var secrets []string
	hostnames := make(map[string]string, 0)
	var gwHostnames []string
	//var hostnames map[string]string
	gwStatus := akogatewayapiobjects.GatewayApiLister().GetGatewayToGatewayStatusMapping(gwNsName)

	for i, listenerObj := range gwObj.Spec.Listeners {
		if akogatewayapilib.IsListenerInvalid(gwStatus, i) {
			continue
		}
		gwListener := akogatewayapiobjects.GatewayListenerStore{}
		gwListener.Name = string(listenerObj.Name)
		gwListener.Gateway = gwNsName
		gwListener.Port = int32(listenerObj.Port)
		gwListener.Protocol = string(listenerObj.Protocol)

		if listenerObj.AllowedRoutes == nil {
			gwListener.AllowedRouteNs = gwObj.Namespace
			gwListener.AllowedRouteTypes = []akogatewayapiobjects.GatewayRouteKind{
				{Group: akogatewayapilib.GatewayGroup, Kind: akogatewayapilib.ProtocolToRoute(gwListener.Protocol)},
			}
		} else {
			if listenerObj.AllowedRoutes.Namespaces != nil {
				if listenerObj.AllowedRoutes.Namespaces.From != nil {
					if string(*listenerObj.AllowedRoutes.Namespaces.From) == akogatewayapilib.AllowedRoutesNamespaceFromSame {
						gwListener.AllowedRouteNs = gwObj.Namespace
					} else if string(*listenerObj.AllowedRoutes.Namespaces.From) == akogatewayapilib.AllowedRoutesNamespaceFromAll {
						gwListener.AllowedRouteNs = akogatewayapilib.AllowedRoutesNamespaceFromAll
					}
				}
			} else {
				gwListener.AllowedRouteNs = gwObj.Namespace
			}
			if listenerObj.AllowedRoutes.Kinds != nil {
				for _, routeKind := range listenerObj.AllowedRoutes.Kinds {
					if routeKind.Group == nil {
						gwListener.AllowedRouteTypes = append(gwListener.AllowedRouteTypes, akogatewayapiobjects.GatewayRouteKind{Group: akogatewayapilib.GatewayGroup, Kind: string(routeKind.Kind)})
					} else {
						if string(*routeKind.Group) == "" {
							gwListener.AllowedRouteTypes = append(gwListener.AllowedRouteTypes, akogatewayapiobjects.GatewayRouteKind{Group: akogatewayapilib.CoreGroup, Kind: string(routeKind.Kind)})
						} else {
							gwListener.AllowedRouteTypes = append(gwListener.AllowedRouteTypes, akogatewayapiobjects.GatewayRouteKind{Group: string(*routeKind.Group), Kind: string(routeKind.Kind)})
						}
					}
				}
			}
		}
		if listenerObj.TLS != nil {
			for _, cert := range listenerObj.TLS.CertificateRefs {
				certNs := gwObj.Namespace
				if cert.Namespace != nil {
					certNs = string(*cert.Namespace)
				}
				secrets = append(secrets, certNs+"/"+string(cert.Name))
			}
		}
		listeners = append(listeners, gwListener)
		if listenerObj.Hostname != nil && string(*listenerObj.Hostname) != "" {
			hostnames[string(listenerObj.Name)] = string(*listenerObj.Hostname)
			gwHostnames = append(gwHostnames, string(*listenerObj.Hostname))
		} else {
			hostnames[string(listenerObj.Name)] = utils.WILDCARD
			gwHostnames = append(gwHostnames, utils.WILDCARD)
		}

	}
	uniqueHostnames := sets.NewString(gwHostnames...)
	//TODO: verify hostname overlap here or use the store updated from here
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToHostnames(gwNsName, uniqueHostnames.List())

	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToListener(gwNsName, listeners)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToSecret(gwNsName, secrets)
	for listener, hostname := range hostnames {
		gwListenerNsName := gwNsName + "/" + listener
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayListenerToHostname(gwListenerNsName, hostname)
	}

	return []string{gwNsName}, true
}

func GatewayToRoutes(namespace, name, key string) ([]string, bool) {
	gatewayObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(string(name))
	if err != nil {
		// does not exist or any other error. do not use it
		if errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: Gateway %s/%s does not exist.", key, namespace, name)
			return nil, false
		}
		utils.AviLog.Errorf("key: %s, msg: Error in fetching gateway details %s/%s. Error: %v", key, namespace, name, err.Error())
		return nil, false
	}
	allowedRoutes := false
	for _, listener := range gatewayObj.Spec.Listeners {
		if listener.AllowedRoutes != nil && listener.AllowedRoutes.Namespaces != nil && listener.AllowedRoutes.Namespaces.From != nil {
			if string(*listener.AllowedRoutes.Namespaces.From) == akogatewayapilib.AllowedRoutesNamespaceFromAll {
				allowedRoutes = true
				break
			}
		}
	}
	routeTypeNsNameList, _ := validateReferredHTTPRoute(key, name, namespace, allowedRoutes)
	return routeTypeNsNameList, true
}

func GatewayClassGetGw(namespace, name, key string) ([]string, bool) {
	var gatewayList []string
	var controllerName string
	isDelete := false
	gwClassObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: got error while getting gateway class: %v", key, err)
			return []string{}, false
		}
		isDelete = true
	} else {
		controllerName = string(gwClassObj.Spec.ControllerName)
	}

	if isDelete {
		gatewayList = akogatewayapiobjects.GatewayApiLister().GetGatewayClassToGateway(name)
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayClass(name)
	} else {
		isAKOController := akogatewayapilib.CheckGatewayClassController(controllerName)
		if isAKOController {
			utils.AviLog.Debugf("key: %s, msg: controller is AKO", key)
		}
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(name, isAKOController)
		gatewayList = akogatewayapiobjects.GatewayApiLister().GetGatewayClassToGateway(name)
	}

	return gatewayList, true
}

func HTTPRouteToGateway(namespace, name, key string) ([]string, bool) {

	routeTypeNsName := lib.HTTPRoute + "/" + namespace + "/" + name
	hrObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		found, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(routeTypeNsName)
		if !found {
			return []string{}, true
		}
		return gwNsNameList, true
	}
	var gwNsNameList []string
	parentNameToHostnameMap := make(map[string][]string)
	gatewayToListenersMap := make(map[string][]akogatewayapiobjects.GatewayListenerStore)
	statusIndex := 0
	httpRouteStatus := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(routeTypeNsName)
	for _, parentRef := range hrObj.Spec.ParentRefs {
		if statusIndex >= len(httpRouteStatus.Parents) {
			break
		}
		if httpRouteStatus.Parents[statusIndex].ParentRef.Name != parentRef.Name {
			continue
		}
		if httpRouteStatus.Parents[statusIndex].Conditions[0].Type == string(gatewayv1.RouteConditionAccepted) && httpRouteStatus.Parents[statusIndex].Conditions[0].Status == metav1.ConditionFalse {
			statusIndex += 1
			continue
		}
		ns := namespace
		if parentRef.Namespace != nil {
			ns = string(*parentRef.Namespace)
			// if *parentRef.Namespace != gatewayv1beta1.Namespace(hrObj.Namespace) {
			// 	//check reference grant
			// }
		}
		// Check gateway present or not
		_, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(ns).Get(string(parentRef.Name))
		if err != nil {
			// does not exist or any other error. do not use it
			if errors.IsNotFound(err) {
				utils.AviLog.Errorf("key: %s, msg: Gateway %s/%s does not exist.", key, ns, parentRef.Name)
				continue
			}
			utils.AviLog.Errorf("key: %s, msg: Error in fetching gateway details %s/%s. Error: %v", key, ns, parentRef.Name, err.Error())
			continue
		}
		parentRefGatewayMappings(parentRef, parentNameToHostnameMap, gatewayToListenersMap, &gwNsNameList, hrObj, namespace, key)
		statusIndex += 1
	}
	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gwNsNameList)
	return gwNsNameList, true
}

func HTTPRouteChanges(namespace, name, key string) ([]string, bool) {
	routeTypeNsName := lib.HTTPRoute + "/" + namespace + "/" + name
	hrObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: got error while getting httproute: %v", key, err)
			return []string{}, false
		}
		// httproute must be deleted so remove mappings
		utils.AviLog.Debugf("Deleting from store")
		//delete route to service must also update gateway to service (through route)
		akogatewayapiobjects.GatewayApiLister().DeleteRouteFromStore(routeTypeNsName, key)
		return []string{routeTypeNsName}, true
	}

	var gwNsNameList []string
	for _, parentRef := range hrObj.Spec.ParentRefs {
		ns := namespace
		if parentRef.Namespace != nil {
			ns = string(*parentRef.Namespace)
		}
		gwNsName := ns + "/" + string(parentRef.Name)
		gwNsNameList = append(gwNsNameList, gwNsName)
	}

	var svcNsNameList []string
	var l7RuleNsNameList []string
	for _, rule := range hrObj.Spec.Rules {
		for _, backendRef := range rule.BackendRefs {
			ns := namespace
			if backendRef.Namespace != nil {
				ns = string(*backendRef.Namespace)
			}
			svcNsName := ns + "/" + string(backendRef.Name)
			svcNsNameList = append(svcNsNameList, svcNsName)
		}
		for _, filter := range rule.Filters {
			// Do we need to check first condition??
			if filter.Type == gatewayv1.HTTPRouteFilterExtensionRef && filter.ExtensionRef != nil {
				if filter.ExtensionRef.Kind == lib.L7Rule {
					l7RuleNsName := namespace + "/" + string(filter.ExtensionRef.Name)
					if !utils.HasElem(l7RuleNsNameList, l7RuleNsName) {
						l7RuleNsNameList = append(l7RuleNsNameList, l7RuleNsName)
					}
				}
			}

		}
	}

	// deletes the services, which are removed, from the gateway <-> service and route <-> service mappings
	found, oldSvcs := akogatewayapiobjects.GatewayApiLister().GetRouteToService(routeTypeNsName)
	if found {
		for _, svcNsName := range oldSvcs {
			if !utils.HasElem(svcNsNameList, svcNsName) {
				akogatewayapiobjects.GatewayApiLister().DeleteRouteToServiceMappings(routeTypeNsName, svcNsName, key)
			}
		}
	}

	// Delete old entries from HTTPRoute->L7RuleMapping & L7Rule-->HTTPRoute
	routeNSName := namespace + "/" + name
	found, oldL7RuleNSNameList := akogatewayapiobjects.GatewayApiLister().GetHTTPRouteToL7RuleMapping(routeNSName)
	if found {
		for l7RuleNsName := range oldL7RuleNSNameList {
			if !utils.HasElem(l7RuleNsNameList, l7RuleNsName) {
				akogatewayapiobjects.GatewayApiLister().DeleteHTTPRouteToL7RuleMapping(routeNSName, l7RuleNsName)
				akogatewayapiobjects.GatewayApiLister().DeleteL7RuleToHTTPRouteMapping(l7RuleNsName, routeNSName)
			}
		}
	}

	// update with new entries for HTTPRoute->L7 Rule and L7Rule to HTTPRoute
	for _, l7RuleNsName := range l7RuleNsNameList {
		akogatewayapiobjects.GatewayApiLister().UpdateHTTPRouteToL7RuleMapping(routeNSName, l7RuleNsName)
		akogatewayapiobjects.GatewayApiLister().UpdateL7RuleToHTTPRouteMapping(l7RuleNsName, routeNSName)
	}

	found, oldGateways := akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(routeTypeNsName)
	if found {
		for _, gwNsName := range oldGateways {
			if !utils.HasElem(gwNsNameList, gwNsName) {
				akogatewayapiobjects.GatewayApiLister().DeleteRouteToGatewayMappings(routeTypeNsName, gwNsName)
			}
		}
	}

	// updates route <-> service mappings with new services
	for _, svcNsName := range svcNsNameList {
		akogatewayapiobjects.GatewayApiLister().UpdateRouteServiceMappings(routeTypeNsName, svcNsName, key)
	}

	// updates gateway <-> service mappings with new services
	for _, gwNsName := range gwNsNameList {
		for _, svcNsName := range svcNsNameList {
			akogatewayapiobjects.GatewayApiLister().UpdateGatewayServiceMappings(gwNsName, svcNsName)
		}
	}

	for _, gwNsName := range oldGateways {
		if utils.HasElem(gwNsNameList, gwNsName) {
			continue
		}
		for _, svcNsName := range oldSvcs {
			if utils.HasElem(svcNsNameList, svcNsName) {
				continue
			}
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayServiceMappings(gwNsName, svcNsName)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: HTTPRoutes retrieved %s", key, []string{routeTypeNsName})
	return []string{routeTypeNsName}, true
}

func ServiceToGateways(namespace, name, key string) ([]string, bool) {
	svcNsName := namespace + "/" + name
	found, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToGateway(svcNsName)
	if !found {
		return []string{}, true
	}
	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gwNsNameList)
	return gwNsNameList, found
}

func ServiceToRoutes(namespace, name, key string) ([]string, bool) {
	svcNsName := namespace + "/" + name
	found, routeTypeNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToRoute(svcNsName)
	if !found {
		return []string{}, true
	}
	utils.AviLog.Debugf("key: %s, msg: Routes retrieved %s", key, routeTypeNsNameList)
	if lib.AutoAnnotateNPLSvc() {
		service, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
		if err != nil || service.Spec.Type == corev1.ServiceTypeNodePort {
			statusOption := status.StatusOptions{
				ObjType:   lib.NPLService,
				Op:        lib.DeleteStatus,
				ObjName:   name,
				Namespace: namespace,
				Key:       key,
			}
			status.PublishToStatusQueue(name, statusOption)
		} else if !status.CheckNPLSvcAnnotation(key, namespace, name) {
			statusOption := status.StatusOptions{
				ObjType:   lib.NPLService,
				Op:        lib.UpdateStatus,
				ObjName:   name,
				Namespace: namespace,
				Key:       key,
			}
			status.PublishToStatusQueue(name, statusOption)
		}
	}
	return routeTypeNsNameList, found
}

func EndpointSlicesToGateways(namespace, name, key string) ([]string, bool) {
	return ServiceToGateways(namespace, name, key)
}

func EndpointSlicesToRoutes(namespace, name, key string) ([]string, bool) {
	return ServiceToRoutes(namespace, name, key)
}

func SecretToGateways(namespace, name, key string) ([]string, bool) {
	secretNsName := namespace + "/" + name
	found, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetSecretToGateway(secretNsName)
	if !found {
		return []string{}, true
	}
	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gwNsNameList)
	return gwNsNameList, found
}

func PodToGateway(namespace, name, key string) ([]string, bool) {
	podNsName := namespace + "/" + name
	pod, err := utils.GetInformers().PodInformer.Lister().Pods(namespace).Get(name)

	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Infof("key: %s, got error while getting pod: %v", key, err)
			return []string{}, false
		}
		utils.AviLog.Debugf("key: %s, msg: Pod not found, mappings will be deleted ", key)
		servicesList := akogatewayapiobjects.GatewayApiLister().GetPodsToService(podNsName)
		gatewayList := []string{}
		for _, svcNsName := range servicesList {
			found, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToGateway(svcNsName)
			if found {
				for _, gwNsName := range gwNsNameList {
					if !utils.HasElem(gatewayList, gwNsName) {
						gatewayList = append(gatewayList, gwNsName)
					}
				}

			}
		}
		return gatewayList, true
	}
	ann := pod.GetAnnotations()
	var annotations []lib.NPLAnnotation
	if err := json.Unmarshal([]byte(ann[lib.NPLPodAnnotation]), &annotations); err != nil {
		utils.AviLog.Warnf("key: %s, got error while unmarshaling NPL annotations: %v", key, err)
	}
	objects.SharedNPLLister().Save(podNsName, annotations)

	servicesList, _ := lib.GetServicesForPod(pod)
	oldServicesList := akogatewayapiobjects.GatewayApiLister().GetPodsToService(podNsName)
	for _, svc := range oldServicesList {
		if !utils.HasElem(servicesList, svc) {
			servicesList = append(servicesList, svc)
		}
	}

	utils.AviLog.Infof("key: %s, msg: NPL Services retrieved: %s", key, servicesList)
	gatewayList := []string{}
	for _, svcNsName := range servicesList {
		found, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToGateway(svcNsName)
		if found {
			for _, gwNsName := range gwNsNameList {
				if !utils.HasElem(gatewayList, gwNsName) {
					gatewayList = append(gatewayList, gwNsName)
				}
			}
		}
	}
	return gatewayList, true

}
func PodToHTTPRoute(namespace, name, key string) ([]string, bool) {
	podNsName := namespace + "/" + name
	pod, err := utils.GetInformers().PodInformer.Lister().Pods(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Infof("key: %s, got error while getting pod: %v", key, err)
			return []string{}, false
		}
		utils.AviLog.Infof("key: %s, msg: Pod not found, deleting mappings", key)

		servicesList := akogatewayapiobjects.GatewayApiLister().GetPodsToService(podNsName)
		akogatewayapiobjects.GatewayApiLister().DeletePodsToService(podNsName)
		objects.SharedNPLLister().Delete(podNsName)
		routeList := []string{}
		for _, serviceNsName := range servicesList {
			found, routeNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToRoute(serviceNsName)
			if found {
				for _, routeNsName := range routeNsNameList {
					if !utils.HasElem(routeList, routeNsName) {
						routeList = append(routeList, routeNsName)
					}
				}
			}
		}
		return routeList, true
	}

	servicesList, _ := lib.GetServicesForPod(pod)
	oldServicesList := akogatewayapiobjects.GatewayApiLister().GetPodsToService(podNsName)
	akogatewayapiobjects.GatewayApiLister().UpdatePodsToService(podNsName, servicesList)
	for _, svc := range oldServicesList {
		if !utils.HasElem(servicesList, svc) {
			servicesList = append(servicesList, svc)
		}
	}
	routeList := []string{}
	for _, serviceNsName := range servicesList {
		found, routeNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToRoute(serviceNsName)
		if found {
			for _, routeNsName := range routeNsNameList {
				if !utils.HasElem(routeList, routeNsName) {
					routeList = append(routeList, routeNsName)
				}
			}
		}
	}
	return routeList, true
}
func NoOperation(namespace, name, key string) ([]string, bool) {
	// No-op
	return []string{}, true
}

func parentRefGatewayMappings(parentRef gatewayv1.ParentReference,
	parentNameToHostnameMap map[string][]string,
	gatewayToListenersMap map[string][]akogatewayapiobjects.GatewayListenerStore,
	gwNsNameList *[]string,
	hrObj *gatewayv1.HTTPRoute,
	namespace, key string) {
	routeTypeNsName := lib.HTTPRoute + "/" + hrObj.Namespace + "/" + hrObj.Name
	httpGroupKind := akogatewayapiobjects.GatewayRouteKind{Group: akogatewayapilib.GatewayGroup, Kind: lib.HTTPRoute}
	hostnameIntersection, _ := parentNameToHostnameMap[string(parentRef.Name)]
	ns := namespace
	if parentRef.Namespace != nil {
		ns = string(*parentRef.Namespace)
		// if *parentRef.Namespace != gatewayv1beta1.Namespace(hrObj.Namespace) {
		// 	//check reference grant
		// }
	}

	var gatewayListenerList []akogatewayapiobjects.GatewayListenerStore
	gwNsName := ns + "/" + string(parentRef.Name)
	listeners := akogatewayapiobjects.GatewayApiLister().GetGatewayToListeners(gwNsName)
	for _, listener := range listeners {
		//check if namespace is allowed
		// TODO: akshay: add selector condition here.
		if (len(listener.AllowedRouteTypes) == 0 || utils.HasElem(listener.AllowedRouteTypes, httpGroupKind)) &&
			(listener.AllowedRouteNs == akogatewayapilib.AllowedRoutesNamespaceFromAll || listener.AllowedRouteNs == hrObj.Namespace) {
			//if provided, check if section name and port matches
			if (parentRef.SectionName == nil || string(*parentRef.SectionName) == listener.Name) &&
				(parentRef.Port == nil || int32(*parentRef.Port) == listener.Port) {

				gwListenerNsName := gwNsName + "/" + listener.Name
				listenerHostname := akogatewayapiobjects.GatewayApiLister().GetGatewayListenerToHostname(gwListenerNsName)

				hostnameMatched := false
				for _, routeHostname := range hrObj.Spec.Hostnames {
					// When Gateway hostname is empty, then just check validity of hostname and append it.
					// When hostname in HTTProute has wildcard
					// When there is exact match
					if listenerHostname == "" || utils.CheckSubdomainOverlapping(string(routeHostname), listenerHostname) {
						if akogatewayapilib.VerifyHostnameSubdomainMatch(string(routeHostname)) {
							hostnameIntersection = append(hostnameIntersection, string(routeHostname))
							hostnameMatched = true
						}
					}
				}
				// If no hostname in HTTPRoute and listener hostname is not empty, include listener hostname
				// into list of mapped hostname between httproute and gateway (empty hostname at route and gateway)
				if len(hrObj.Spec.Hostnames) == 0 && (listenerHostname != "" && listenerHostname != "*") {
					hostnameIntersection = append(hostnameIntersection, string(listenerHostname))
				}

				if (hostnameMatched && !utils.HasElem(gatewayListenerList, listener)) || (len(hrObj.Spec.Hostnames) == 0 && (listenerHostname != "" && listenerHostname != "*")) {
					gatewayListenerList = append(gatewayListenerList, listener)
				}
			}
		}
	}

	if len(gatewayListenerList) > 0 {
		gatewayToListenersMapList, _ := gatewayToListenersMap[gwNsName]
		for _, gwListener := range gatewayListenerList {
			if !utils.HasElem(gatewayToListenersMapList, gwListener) {
				gatewayToListenersMapList = append(gatewayToListenersMapList, gwListener)
			}
		}
		gatewayToListenersMap[gwNsName] = gatewayToListenersMapList
	}
	uniqueHosts := sets.NewString(hostnameIntersection...)
	gwRouteNsName := fmt.Sprintf("%s/%s", gwNsName, routeTypeNsName)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteToHostname(gwRouteNsName, uniqueHosts.List())
	utils.AviLog.Infof("key: %s, msg: Hosts mapped to GatewayRoute [%s] are [%v]", key, gwRouteNsName, uniqueHosts.List())
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteMappings(gwNsName, routeTypeNsName)
	utils.AviLog.Infof("key: %s, msg: Routes mapped to Gateway [%v] are : [%v]", key, gwNsName, routeTypeNsName)
	akogatewayapiobjects.GatewayApiLister().UpdateRouteToGatewayListenerMappings(gatewayToListenersMap[gwNsName], routeTypeNsName, gwNsName)
	if !utils.HasElem(gwNsNameList, gwNsName) {
		*gwNsNameList = append(*gwNsNameList, gwNsName)
	}
	parentNameToHostnameMap[string(parentRef.Name)] = hostnameIntersection
}

func validateReferredHTTPRoute(key, name, namespace string, allowedRoutesAll bool) ([]string, error) {
	ns := namespace
	if allowedRoutesAll {
		ns = metav1.NamespaceAll
	}
	hrObjs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(ns).List(labels.Set(nil).AsSelector())
	if err != nil {
		return nil, err
	}
	httpRoutes := make([]*gatewayv1.HTTPRoute, 0)
	for _, httpRoute := range hrObjs {
		httpRouteStatus := httpRoute.Status.DeepCopy()
		httpRouteStatus.Parents = make([]gatewayv1.RouteParentStatus, 0, len(httpRoute.Spec.ParentRefs))
		routeTypeNsName := lib.HTTPRoute + "/" + httpRoute.Namespace + "/" + httpRoute.Name
		httpRouteStatusInCache := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(routeTypeNsName)
		if httpRouteStatusInCache == nil {
			continue
		}
		parentRefIndexInHttpRouteStatus := 0
		indexInCache := 0
		appendRoute := false
		for parentRefIndexFromSpec, parentRef := range httpRoute.Spec.ParentRefs {
			matchNamespace := httpRoute.Namespace
			if parentRef.Namespace != nil {
				matchNamespace = string(*parentRef.Namespace)
			}
			if (parentRef.Name == gatewayv1.ObjectName(name)) && (matchNamespace == namespace) {
				isValidHttprouteRules := validateHTTPRouteRules(key, httpRoute, httpRouteStatus)
				if isValidHttprouteRules {
					err := validateParentReference(key, httpRoute, httpRouteStatus, parentRefIndexFromSpec, &parentRefIndexInHttpRouteStatus, &indexInCache)
					if err != nil {
						parentRefName := parentRef.Name
						utils.AviLog.Warnf("key: %s, msg: Parent Reference %s of HTTPRoute object %s is not valid, err: %v", key, parentRefName, httpRoute.Name, err)
					} else {
						appendRoute = true
					}
				} else {
					utils.AviLog.Warnf("key: %s, msg: HTTPUrlRewrite PathType has Unsupported value.", key)
					appendRoute = false
				}
			} else {
				gwName := parentRef.Name
				namespace := httpRoute.Namespace
				if parentRef.Namespace != nil {
					namespace = string(*parentRef.Namespace)
				}
				gateway, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(string(gwName))
				if err != nil {
					utils.AviLog.Errorf("key: %s, msg: unable to get the gateway object %s . err: %s", key, gwName, err)
					continue
				}
				gwClass := string(gateway.Spec.GatewayClassName)
				_, isAKOCtrl := akogatewayapiobjects.GatewayApiLister().IsGatewayClassControllerAKO(gwClass)
				if !isAKOCtrl {
					utils.AviLog.Warnf("key: %s, msg: controller for the parent reference %s of HTTPRoute object %s is not ako", key, name, httpRoute.Name)
				} else {
					httpRouteStatus.Parents = append(httpRouteStatus.Parents, httpRouteStatusInCache.Parents[indexInCache])
				}
			}
		}

		akogatewayapistatus.Record(key, httpRoute, &status.Status{HTTPRouteStatus: httpRouteStatus})
		if appendRoute {
			httpRoutes = append(httpRoutes, httpRoute)
		}
	}
	sort.Slice(httpRoutes, func(i, j int) bool {
		if httpRoutes[i].GetCreationTimestamp().Unix() == httpRoutes[j].GetCreationTimestamp().Unix() {
			return httpRoutes[i].Namespace+"/"+httpRoutes[i].Name < httpRoutes[j].Namespace+"/"+httpRoutes[j].Name
		}
		return httpRoutes[i].GetCreationTimestamp().Unix() < httpRoutes[j].GetCreationTimestamp().Unix()
	})
	var routes []string
	for _, httpRoute := range httpRoutes {
		httpRouteToGatewayOperation(httpRoute, key, name, namespace)
		routeTypeNsName := lib.HTTPRoute + "/" + httpRoute.Namespace + "/" + httpRoute.Name
		routes = append(routes, routeTypeNsName)
	}
	return routes, nil
}

func httpRouteToGatewayOperation(hrObj *gatewayv1.HTTPRoute, key, gwName, gwNamespace string) {
	routeTypeNsName := lib.HTTPRoute + "/" + hrObj.Namespace + "/" + hrObj.Name
	var gwNsNameList []string
	parentNameToHostnameMap := make(map[string][]string)
	gatewayToListenersMap := make(map[string][]akogatewayapiobjects.GatewayListenerStore)
	statusIndex := 0
	httpRouteStatus := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(routeTypeNsName)
	for _, parentRef := range hrObj.Spec.ParentRefs {
		if statusIndex >= len(httpRouteStatus.Parents) {
			break
		}
		if httpRouteStatus.Parents[statusIndex].ParentRef.Name != parentRef.Name {
			continue
		}
		if httpRouteStatus.Parents[statusIndex].Conditions[0].Type == string(gatewayv1.RouteConditionAccepted) && httpRouteStatus.Parents[statusIndex].Conditions[0].Status == metav1.ConditionFalse {
			statusIndex += 1
			continue
		}
		if string(parentRef.Name) != gwName {
			statusIndex += 1
			continue
		}
		if parentRef.Namespace != nil && string(*parentRef.Namespace) != gwNamespace {
			statusIndex += 1
			continue
		}
		parentRefGatewayMappings(parentRef, parentNameToHostnameMap, gatewayToListenersMap, &gwNsNameList, hrObj, gwNamespace, key)
		statusIndex += 1
	}
	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gwNsNameList)
}

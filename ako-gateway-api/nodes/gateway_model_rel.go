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
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
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
	Endpoint = GraphSchema{
		Type:        "Endpoints",
		GetGateways: EndpointToGateways,
		GetRoutes:   EndpointToRoutes,
	}
	EndpointSlices = GraphSchema{
		Type:        utils.Endpointslices,
		GetGateways: EndpointToGateways,
		GetRoutes:   EndpointToRoutes,
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
		Endpoint,
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
		if IsListenerInvalid(gwStatus, i) {
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
	gwNsName := namespace + "/" + name
	found, routeTypeNsNameList := akogatewayapiobjects.GatewayApiLister().GetGatewayToRoute(gwNsName)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: No route objects mapped to this gateway", key)
		return []string{}, true
	}
	return routeTypeNsNameList, found
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
	httpGroupKind := akogatewayapiobjects.GatewayRouteKind{Group: akogatewayapilib.GatewayGroup, Kind: lib.HTTPRoute}
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
	var listenerList []akogatewayapiobjects.GatewayListenerStore
	var gatewayList []string
	var gwNsNameList []string
	parentNameToHostnameMap := make(map[string][]string)
	statusIndex := 0
	httpRouteStatus := akogatewayapiobjects.GatewayApiLister().GetRouteToRouteStatusMapping(routeTypeNsName)
	for _, parentRef := range hrObj.Spec.ParentRefs {
		if statusIndex >= len(httpRouteStatus.Parents) {
			break
		}
		if httpRouteStatus.Parents[statusIndex].ParentRef.Name != parentRef.Name {
			continue
		}
		for statusIndex < len(httpRouteStatus.Parents) && (parentRef.SectionName != nil && *httpRouteStatus.Parents[statusIndex].ParentRef.SectionName != *parentRef.SectionName) {
			statusIndex += 1
		}
		if httpRouteStatus.Parents[statusIndex].Conditions[0].Type == string(gatewayv1.RouteConditionAccepted) && httpRouteStatus.Parents[statusIndex].Conditions[0].Status == metav1.ConditionFalse {
			continue
		}
		hostnameIntersection, _ := parentNameToHostnameMap[string(parentRef.Name)]
		ns := namespace
		if parentRef.Namespace != nil {
			ns = string(*parentRef.Namespace)
			// if *parentRef.Namespace != gatewayv1beta1.Namespace(hrObj.Namespace) {
			// 	//check reference grant
			// }
		}

		var gatewayListenerList []akogatewayapiobjects.GatewayListenerStore

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
			if !utils.HasElem(gatewayList, gwNsName) {
				gatewayList = append(gatewayList, gwNsName)
			}
			for _, gwListener := range gatewayListenerList {
				if !utils.HasElem(listenerList, gwListener) {
					listenerList = append(listenerList, gwListener)
				}
			}
		}
		uniqueHosts := sets.NewString(hostnameIntersection...)
		gwRouteNsName := fmt.Sprintf("%s/%s", gwNsName, routeTypeNsName)
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteToHostname(gwRouteNsName, uniqueHosts.List())
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteMappings(gwNsName, routeTypeNsName)
		if !utils.HasElem(gwNsNameList, gwNsName) {
			gwNsNameList = append(gwNsNameList, gwNsName)
		}
		parentNameToHostnameMap[string(parentRef.Name)] = hostnameIntersection
		statusIndex += 1
	}
	akogatewayapiobjects.GatewayApiLister().UpdateRouteToGatewayListenerMappings(listenerList, routeTypeNsName)
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

		//delete route to service must also update gateway to service (through route)
		akogatewayapiobjects.GatewayApiLister().DeleteRouteFromStore(routeTypeNsName)
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
	for _, rule := range hrObj.Spec.Rules {
		for _, backendRef := range rule.BackendRefs {
			ns := namespace
			if backendRef.Namespace != nil {
				ns = string(*backendRef.Namespace)
			}
			svcNsName := ns + "/" + string(backendRef.Name)
			svcNsNameList = append(svcNsNameList, svcNsName)
		}
	}

	// deletes the services, which are removed, from the gateway <-> service and route <-> service mappings
	found, oldSvcs := akogatewayapiobjects.GatewayApiLister().GetRouteToService(routeTypeNsName)
	if found {
		for _, svcNsName := range oldSvcs {
			if !utils.HasElem(svcNsNameList, svcNsName) {
				akogatewayapiobjects.GatewayApiLister().DeleteRouteToServiceMappings(routeTypeNsName, svcNsName)
			}
		}
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
		akogatewayapiobjects.GatewayApiLister().UpdateRouteServiceMappings(routeTypeNsName, svcNsName)
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
	return routeTypeNsNameList, found
}

func EndpointToGateways(namespace, name, key string) ([]string, bool) {
	return ServiceToGateways(namespace, name, key)
}

func EndpointToRoutes(namespace, name, key string) ([]string, bool) {
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

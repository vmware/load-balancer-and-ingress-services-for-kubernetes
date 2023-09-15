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
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
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
	HTTPRoute = GraphSchema{
		Type:        lib.HTTPRoute,
		GetGateways: HTTPRouteToGateway,
		GetRoutes:   HTTPRouteChanges,
	}
	SupportedGraphTypes = GraphDescriptor{
		Gateway,
		GatewayClass,
		Secret,
		Service,
		Endpoint,
		HTTPRoute,
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
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayToSecret(gwNsName)
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayToRoute(gwNsName)
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayToGatewayClass(gwNsName)
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayToService(gwNsName)

		return []string{gwNsName}, true
	}

	gwClassName := string(gwObj.Spec.GatewayClassName)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToGatewayClass(namespace, name, gwClassName)

	var listeners []string
	var secrets []string
	hostnames := make(map[string]string, 0)
	//var hostnames map[string]string
	for _, listenerObj := range gwObj.Spec.Listeners {
		listenerString := string(listenerObj.Name) + "/" +
			strconv.Itoa(int(listenerObj.Port)) + "/" + string(listenerObj.Protocol)
		if listenerObj.AllowedRoutes == nil {
			listenerString += "/" + string(gwObj.Namespace)
		} else {
			if listenerObj.AllowedRoutes.Namespaces != nil {
				if listenerObj.AllowedRoutes.Namespaces.From != nil {
					if string(*listenerObj.AllowedRoutes.Namespaces.From) == "Same" {
						listenerString += "/" + string(gwObj.Namespace)
					} else if string(*listenerObj.AllowedRoutes.Namespaces.From) == "All" {
						listenerString += "/All"
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
		listeners = append(listeners, listenerString)
		hostnames[string(listenerObj.Name)] = string(*listenerObj.Hostname)
	}
	sort.Strings(listeners)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToListener(gwNsName, listeners)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToSecret(gwNsName, secrets)
	for listener, hostname := range hostnames {
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayListenerToHostname(namespace+"/"+name+"/"+listener, hostname)
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
	var listenerList []string
	var gatewayList []string
	var hostnameIntersection []string
	var gwNsNameList []string
	for _, parentRef := range hrObj.Spec.ParentRefs {
		ns := namespace
		if parentRef.Namespace != nil {
			ns = string(*parentRef.Namespace)
			// if *parentRef.Namespace != gatewayv1beta1.Namespace(hrObj.Namespace) {
			// 	//check reference grant
			// }
		}
		var gatewayListenerList []string
		gwNsName := ns + "/" + string(parentRef.Name)
		listeners := akogatewayapiobjects.GatewayApiLister().GetGatewayToListeners(gwNsName)
		for _, listener := range listeners {
			listenerSlice := strings.Split(listener, "/")
			listenerName := listenerSlice[0]
			listenerPort := listenerSlice[1]
			listenerAllowedNS := listenerSlice[3]
			//check if namespace is allowed
			if listenerAllowedNS == "All" || listenerAllowedNS == hrObj.Namespace {
				//if provided, check if section name and port matches
				if (parentRef.SectionName == nil || string(*parentRef.SectionName) == listenerName) &&
					(parentRef.Port == nil || string(*parentRef.Port) == listenerPort) {
					listenerHostname := akogatewayapiobjects.GatewayApiLister().GetGatewayListenerToHostname(gwNsName, listenerName)
					if strings.HasPrefix(listenerHostname, "*") {
						listenerHostname = listenerHostname[1:]
					}
					hostnameMatched := false
					for _, routeHostname := range hrObj.Spec.Hostnames {
						if strings.HasSuffix(string(routeHostname), listenerHostname) {
							hostnameIntersection = append(hostnameIntersection, string(routeHostname))
							hostnameMatched = true
						}
					}
					if hostnameMatched && !utils.HasElem(listenerList, gwNsName+"/"+listenerName) {
						gatewayListenerList = append(listenerList, gwNsName+"/"+listenerName)
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
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteToHostname(gwNsName, hostnameIntersection)

		akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteMappings(gwNsName, listenerList, routeTypeNsName)
		if !utils.HasElem(gwNsNameList, gwNsName) {
			gwNsNameList = append(gwNsNameList, gwNsName)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gwNsNameList)
	return gwNsNameList, true
}

func HTTPRouteChanges(namespace, name, key string) ([]string, bool) {
	routeTypeNsName := lib.HTTPRoute + "/" + namespace + "/" + name
	hrObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		_, svcNsNameList := akogatewayapiobjects.GatewayApiLister().GetRouteToService(routeTypeNsName)
		_, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(routeTypeNsName)
		for _, gwNsName := range gwNsNameList {
			for _, svcNsName := range svcNsNameList {
				akogatewayapiobjects.GatewayApiLister().DeleteGatewayServiceMappings(gwNsName, svcNsName)
			}
		}
		akogatewayapiobjects.GatewayApiLister().DeleteRouteServiceMappings(routeTypeNsName)
		akogatewayapiobjects.GatewayApiLister().DeleteRouteGatewayMappings(routeTypeNsName)
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
				for _, gwNsName := range gwNsNameList {
					akogatewayapiobjects.GatewayApiLister().DeleteGatewayServiceMappings(gwNsName, svcNsName)
				}
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

func NoOperation(namespace, name, key string) ([]string, bool) {
	// No-op
	return []string{}, true
}

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

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func ConfigDescriptor() GraphDescriptor {
	return SupportedGraphTypes
}

type GraphDescriptor []GraphSchema

type GraphSchema struct {
	Type        string
	GetGateways func(string, string, string) ([]string, bool)
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
	}
	GatewayClass = GraphSchema{
		Type:        "GatewayClass",
		GetGateways: GatewayClassGetGw,
	}
	HTTPRoute = GraphSchema{
		Type:        lib.HTTPRoute,
		GetGateways: HTTPRouteToGateway,
	}
	Service = GraphSchema{
		Type:        "Service",
		GetGateways: ServiceToGateway,
	}
	SupportedGraphTypes = GraphDescriptor{
		Gateway,
		GatewayClass,
		HTTPRoute,
		Service,
	}
)

func GatewayGetGw(namespace, name, key string) ([]string, bool) {
	gw := []string{namespace + "/" + name}

	gwObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayInformer.Lister().Gateways(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		return gw, true
	}
	gwClassName := string(gwObj.Spec.GatewayClassName)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToGatewayClass(namespace, name, gwClassName)
	var listeners []string
	hostnames := make(map[string]string, 0)
	//var hostnames map[string]string
	for _, l := range gwObj.Spec.Listeners {
		s := string(l.Name) + "/" + strconv.Itoa(int(l.Port)) + "/" + string(l.Protocol)
		if l.AllowedRoutes == nil {
			s += "/" + string(gwObj.Namespace)
		} else {
			if l.AllowedRoutes.Namespaces != nil {
				if l.AllowedRoutes.Namespaces.From != nil {
					if string(*l.AllowedRoutes.Namespaces.From) == "Same" {
						s += "/" + string(gwObj.Namespace)
					} else if string(*l.AllowedRoutes.Namespaces.From) == "All" {
						s += "/All"
					}
				}
			}
		}
		listeners = append(listeners, s)
		hostnames[string(l.Name)] = string(*l.Hostname)
	}
	sort.Strings(listeners)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToListener(namespace+"/"+name, listeners)
	for listener, hostname := range hostnames {
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayListenerToHostname(namespace+"/"+name+"/"+listener, hostname)
	}

	return gw, true
}

func GatewayClassGetGw(namespace, name, key string) ([]string, bool) {
	var controllerName string
	isDelete := false
	gwClassObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, got error while getting gateway class: %v", key, err)
			return []string{}, false
		}
		isDelete = true
	} else {
		controllerName = string(gwClassObj.Spec.ControllerName)
	}

	if isDelete {
		akogatewayapiobjects.GatewayApiLister().DeleteGatewayClass(name)
	} else {
		isAKOController := akogatewayapilib.CheckGatewayClassController(controllerName)
		if isAKOController {
			utils.AviLog.Debugf("key: %s, controller is AKO", key)
		}
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayClass(name, isAKOController)
	}
	return akogatewayapiobjects.GatewayApiLister().GetGatewayClassToGateway(name), true
}

func HTTPRouteToGateway(namespace, name, key string) ([]string, bool) {
	var isDeleteCase bool
	httpRouteObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		isDeleteCase = true
	}
	//var gwNsNames []string
	if isDeleteCase {
		//_, gwNsNames = akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(lib.HTTPRoute, namespace+"/"+name)
	}
	var listenerList []string
	var gatewayList []string
	var hostnameIntersection []string
	for _, parent := range httpRouteObj.Spec.ParentRefs {
		parentNsName := string(*parent.Namespace) + "/" + string(parent.Name)
		//check gateway is present in store

		if !akogatewayapiobjects.GatewayApiLister().IsGatewayInStore(parentNsName) {
			continue
		}
		if *parent.Namespace != gatewayv1beta1.Namespace(httpRouteObj.Namespace) {
			//check reference grant
		}
		var gatewayListenerList []string
		listeners := akogatewayapiobjects.GatewayApiLister().GetGatewayToListeners(string(*parent.Namespace), string(parent.Name))
		for _, listener := range listeners {
			listenerSlice := strings.Split(listener, "/")
			listenerName := listenerSlice[0]
			listenerPort := listenerSlice[1]
			listenerAllowedNS := listenerSlice[3]
			//check if namespace is allowed
			if listenerAllowedNS == "All" || listenerAllowedNS == httpRouteObj.Namespace {
				//if provided, check if section name and port matches
				if (parent.SectionName == nil || string(*parent.SectionName) == listenerName) &&
					(parent.Port == nil || string(*parent.Port) == listenerPort) {
					listenerHostname := akogatewayapiobjects.GatewayApiLister().GetGatewayListenerToHostname(string(*parent.Namespace), string(parent.Name), listenerName)
					if strings.HasPrefix(listenerHostname, "*") {
						listenerHostname = listenerHostname[1:]
					}
					hostnameMatched := false
					for _, routeHostname := range httpRouteObj.Spec.Hostnames {
						if strings.HasSuffix(string(routeHostname), listenerHostname) {
							hostnameIntersection = append(hostnameIntersection, string(routeHostname))
							hostnameMatched = true
						}
					}
					if hostnameMatched && !utils.HasElem(listenerList, string(*parent.Namespace)+"/"+string(parent.Name)+"/"+listenerName) {
						gatewayListenerList = append(listenerList, string(*parent.Namespace)+"/"+string(parent.Name)+"/"+listenerName)
					}
				}
			}
		}

		if len(gatewayListenerList) > 0 {
			if !utils.HasElem(gatewayList, string(*parent.Namespace)+"/"+string(parent.Name)) {
				gatewayList = append(gatewayList, string(*parent.Namespace)+"/"+string(parent.Name))
			}
			for _, gwListener := range gatewayListenerList {
				if !utils.HasElem(listenerList, gwListener) {
					listenerList = append(listenerList, gwListener)
				}
			}
		}

		routeNsName := httpRouteObj.Namespace + "/" + httpRouteObj.Name
		if isDeleteCase {
			akogatewayapiobjects.GatewayApiLister().DeleteGatewayListenerRouteMappings("HTTPRoute", routeNsName)
		} else {
			//update all gateways with route
			routeNsName := httpRouteObj.Namespace + "/" + httpRouteObj.Name
			for _, gwListener := range listenerList {
				akogatewayapiobjects.GatewayApiLister().UpdateGatewayListenerRouteMappings(gwListener, "HTTPRoute", routeNsName)
			}
			akogatewayapiobjects.GatewayApiLister().UpdateRouteToGateway(routeNsName, gatewayList)
			akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteToHostname(string(*parent.Namespace), string(parent.Name), hostnameIntersection)
		}
	}

	// found, gwNsNames := akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(lib.HTTPRoute, namespace+"/"+name)
	for _, rule := range httpRouteObj.Spec.Rules {
		if rule.BackendRefs != nil {
			for _, backend := range rule.BackendRefs {
				if backend.BackendObjectReference.Kind != nil && string(*backend.BackendObjectReference.Kind) == "Service" {
					var serviceNsName string
					if backend.Namespace == nil {
						serviceNsName = httpRouteObj.Namespace + "/" + string(backend.Name)
					} else {
						serviceNsName = string(*backend.Namespace) + "/" + string(backend.Name)
					}
					akogatewayapiobjects.GatewayApiLister().UpdateServiceToGateway(serviceNsName, gatewayList)
				}
			}
		}
	}
	return gatewayList, true
}

func ServiceToGateway(namespace, name, key string) ([]string, bool) {
	serviceNsName := namespace + "/" + name
	return akogatewayapiobjects.GatewayApiLister().GetServiceToGateway(serviceNsName)
}

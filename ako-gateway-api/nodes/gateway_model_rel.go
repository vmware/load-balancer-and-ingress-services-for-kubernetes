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
		Type:        "Endpoint",
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
			utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		return []string{gwNsName}, true
	}
	gwClassName := string(gwObj.Spec.GatewayClassName)
	akogatewayapiobjects.GatewayApiLister().UpdateGatewayToGatewayClass(namespace, name, gwClassName)
	return []string{gwNsName}, true
}

func GatewayToRoutes(namespace, name, key string) ([]string, bool) {
	gwNsName := namespace + "/" + name
	found, routeTypeNsNameList := akogatewayapiobjects.GatewayApiLister().GetGatewayToRoute(gwNsName)
	if !found {
		utils.AviLog.Debugf("key: %s, No route objects mapped to this gateway", key)
		return []string{}, true
	}
	return routeTypeNsNameList, found
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
	routeTypeNsName := lib.HTTPRoute + "/" + namespace + "/" + name
	hrObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		found, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(routeTypeNsName)
		if !found {
			return []string{}, true
		}
		akogatewayapiobjects.GatewayApiLister().DeleteRouteGatewayMappings(routeTypeNsName)
		return gwNsNameList, true
	}

	var gwNsNameList []string
	for _, parentRef := range hrObj.Spec.ParentRefs {
		ns := namespace
		if parentRef.Namespace != nil {
			ns = string(*parentRef.Namespace)
		}
		gwNsName := ns + "/" + string(parentRef.Name)
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayRouteMappings(gwNsName, routeTypeNsName)
		gwNsNameList = append(gwNsNameList, gwNsName)
	}

	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gwNsNameList)
	return gwNsNameList, true
}

func HTTPRouteChanges(namespace, name, key string) ([]string, bool) {
	routeTypeNsName := lib.HTTPRoute + "/" + namespace + "/" + name
	hrObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		akogatewayapiobjects.GatewayApiLister().DeleteRouteServiceMappings(routeTypeNsName)
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
			akogatewayapiobjects.GatewayApiLister().UpdateRouteServiceMappings(routeTypeNsName, svcNsName)
		}
	}

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
	var svcDeleted bool
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		svcDeleted = true
	}
	found, gwNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToGateway(svcNsName)
	if !found {
		return []string{}, true
	}
	if svcDeleted {
		akogatewayapiobjects.GatewayApiLister().DeleteServiceGatewayMappings(svcNsName)
	}
	utils.AviLog.Debugf("key: %s, msg: Gateways retrieved %s", key, gwNsNameList)
	return gwNsNameList, found
}

func ServiceToRoutes(namespace, name, key string) ([]string, bool) {
	svcNsName := namespace + "/" + name
	var svcDeleted bool
	_, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
			return []string{}, false
		}
		svcDeleted = true
	}
	found, routeTypeNsNameList := akogatewayapiobjects.GatewayApiLister().GetServiceToRoute(namespace + "/" + name)
	if !found {
		return []string{}, true
	}
	if svcDeleted {
		akogatewayapiobjects.GatewayApiLister().DeleteServiceRouteMappings(svcNsName)
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
	return []string{}, true
}

func NoOperation(namespace, name, key string) ([]string, bool) {
	// No-op
	return []string{}, true
}

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
	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
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
	SupportedGraphTypes = GraphDescriptor{
		Gateway,
		GatewayClass,
		HTTPRoute,
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
	// var isDeleteCase bool
	// httpRouteObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	// if err != nil {
	// 	if !errors.IsNotFound(err) {
	// 		utils.AviLog.Errorf("key: %s, got error while getting gateway: %v", key, err)
	// 		return []string{}, false
	// 	}
	// 	isDeleteCase = true
	// }
	// var gwNsNames []string
	// if isDeleteCase {
	// 	_, gwNsNames = akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(lib.HTTPRoute, namespace+"/"+name)
	// }

	// for _, parent := range httpRouteObj.Spec.ParentRefs {
	// 	_ = namespace
	// 	if parent.Namespace != nil {
	// 		_ = string(*parent.Namespace)
	// 	}
	// 	if isDeleteCase {
	// 		//akogatewayapiobjects.GatewayApiLister().DeleteGatewayToRoute(ns+"/"+string(parent.Name), lib.HTTPRoute, namespace+"/"+name)
	// 	} else {
	// 		//akogatewayapiobjects.GatewayApiLister().UpdateGatewayToRoute(ns+"/"+string(parent.Name), lib.HTTPRoute, namespace+"/"+name)
	// 	}
	// }

	// found, gwNsNames := akogatewayapiobjects.GatewayApiLister().GetRouteToGateway(lib.HTTPRoute, namespace+"/"+name)
	return []string{"default/example-gateway"}, true
}

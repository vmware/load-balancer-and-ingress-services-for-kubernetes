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

package objects

import (
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var gwLister *GWLister
var gwonce sync.Once

func GatewayApiLister() *GWLister {
	gwonce.Do(func() {
		gwLister = &GWLister{
			gatewayClassStore:          objects.NewObjectMapStore(),
			gatewayToGatewayClassStore: objects.NewObjectMapStore(),
			gatewayClassToGatewayStore: objects.NewObjectMapStore(),
		}
	})
	return gwLister
}

type GWLister struct {
	gwLock sync.RWMutex

	//Gateways with AKO as controller
	gatewayClassStore *objects.ObjectMapStore

	//Namespace/Gateway -> GatewayClass
	gatewayToGatewayClassStore *objects.ObjectMapStore

	//GatewayClass -> [ns1/gateway1, ns2/gateway2, ...]
	gatewayClassToGatewayStore *objects.ObjectMapStore

	// namespace/gateway -> [listener/routeType/routeNs/routeName, ...]
	gatewayToRoute *objects.ObjectMapStore

	// routeType/routeNs/routeName -> [namespace/gateway/listener, ...]
	routeToGateway *objects.ObjectMapStore

	//svc -> gw
	//route <-> gw
	//secret -> gw
}

func (g *GWLister) IsGatewayClassControllerAKO(gwClass string) (bool, bool) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	found, isAkoCtrl := g.gatewayClassStore.Get(gwClass)
	if found {
		return true, isAkoCtrl.(bool)
	}
	return false, false
}

func (g *GWLister) UpdateGatewayClass(gwClass string, isAkoCtrl bool) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	found, _ := g.gatewayClassToGatewayStore.Get(gwClass)
	if !found {
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass, make([]string, 0))
	}
	g.gatewayClassStore.AddOrUpdate(gwClass, isAkoCtrl)
}

func (g *GWLister) DeleteGatewayClass(gwClass string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayClassStore.Delete(gwClass)
}

func (g *GWLister) GetGatewayClassToGateway(gwClass string) []string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	found, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	if !found {
		return make([]string, 0)
	}
	return gatewayList.([]string)
}

func (g *GWLister) GetGatewayToGatewayClass(ns, gw string) string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	key := getKeyForGateway(ns, gw)
	_, gwClass := g.gatewayToGatewayClassStore.Get(key)
	return gwClass.(string)
}

func (g *GWLister) UpdateGatewayToGatewayClass(ns, gw, gwClass string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	key := getKeyForGateway(ns, gw)

	//remove gateway from old class list
	if found, oldGwClass := g.gatewayToGatewayClassStore.Get(key); found {
		oldGwClassObj := oldGwClass.(string)
		if ok, gatewayList := g.gatewayClassToGatewayStore.Get(oldGwClassObj); ok {
			gatewayListObj := gatewayList.([]string)
			utils.Remove(gatewayListObj, gw)
			g.gatewayClassToGatewayStore.AddOrUpdate(oldGwClassObj, gatewayListObj)
		}
	}
	g.gatewayToGatewayClassStore.AddOrUpdate(key, gwClass)
	found, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	if !found {
		gatewayList = make([]string, 0)
	}
	gatewayListObj := gatewayList.([]string)
	if !utils.HasElem(gatewayListObj, key) {
		gatewayListObj = append(gatewayListObj, key)
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass, gatewayListObj)
	}
}

func (g *GWLister) DeleteGatewayToGatewayClass(ns, gw string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	key := getKeyForGateway(ns, gw)
	found, gwClass := g.gatewayToGatewayClassStore.Get(key)
	if found {
		g.gatewayToGatewayClassStore.Delete(key)
		_, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass.(string))
		gatewayListObj := gatewayList.([]string)
		utils.Remove(gatewayListObj, key)
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass.(string), gatewayListObj)
	}
}

func getKeyForGateway(ns, gw string) string {
	return ns + "/" + gw
}

func (g *GWLister) GetGatewayToRoute(gwNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.gatewayToRoute.Get(gwNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) DeleteGatewayToRoute(gwNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayToRoute.Delete(gwNsName)
}

func (g *GWLister) GetRouteToGateway(routeType, routeNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.routeToGateway.Get(routeType + "/" + routeNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) DeleteRouteToGateway(routeType, routeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.routeToGateway.Delete(routeType + "/" + routeNsName)
}

func (g *GWLister) UpdateGatewayRouteMappings(gwNsName, routeType, routeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// update gw to route mapping
	if found, obj := g.gatewayToRoute.Get(gwNsName); found {
		routes := obj.([]string)
		routes = append(routes, routeType+"/"+routeNsName)
		g.gatewayToRoute.AddOrUpdate(gwNsName, routes)
	}

	// update route to gw mapping
	routeKey := routeType + "/" + routeNsName
	if found, obj := g.routeToGateway.Get(routeKey); found {
		gwNsNames := obj.([]string)
		gwNsNames = append(gwNsNames, gwNsName)
		g.routeToGateway.AddOrUpdate(routeKey, gwNsNames)
	}
}

func (g *GWLister) DeleteGatewayRouteMappings(gwNsName, routeType, routeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// delete gw to route mapping
	if found, obj := g.gatewayToRoute.Get(gwNsName); found {
		routes := obj.([]string)
		utils.Remove(routes, routeType+"/"+routeNsName)
		g.gatewayToRoute.AddOrUpdate(gwNsName, routes)
	}

	// delete route to gw mapping
	routeKey := routeType + "/" + routeNsName
	if found, obj := g.routeToGateway.Get(routeKey); found {
		gwNsNames := obj.([]string)
		gwNsNames = append(gwNsNames, gwNsName)
		utils.Remove(gwNsNames, gwNsName)
		g.routeToGateway.AddOrUpdate(routeKey, gwNsNames)
	}

}

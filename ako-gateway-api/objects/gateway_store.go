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
	"strings"
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
			gatewayToListenerStore:     objects.NewObjectMapStore(),
			gatewayListenerToHostname:  objects.NewObjectMapStore(),
			gatewayListenerToRoute:     objects.NewObjectMapStore(),
			routeToGatewayListener:     objects.NewObjectMapStore(),
			routeToGateway:             objects.NewObjectMapStore(),
			gatewayRouteToHostname:     objects.NewObjectMapStore(),
			serviceToGateway:           objects.NewObjectMapStore(),
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

	//Namespace/Gateway -> [listener1, listener2, ...]
	gatewayToListenerStore *objects.ObjectMapStore

	//Namespace/Gateway/Listner -> hostname
	gatewayListenerToHostname *objects.ObjectMapStore

	// namespace/gateway/listener -> routeType/routeNs/routeName
	gatewayListenerToRoute *objects.ObjectMapStore

	// routeType/routeNs/routeName -> [namespace/gateway/listener, ...]
	routeToGatewayListener *objects.ObjectMapStore

	//routeType/routeNS/routeName -> [namespace/gateway, ...]
	routeToGateway *objects.ObjectMapStore

	//gatewayNS/gatewayName -> [hostname, ...]
	gatewayRouteToHostname *objects.ObjectMapStore

	//namespace/service -> [namespace/gateway, ...]
	serviceToGateway *objects.ObjectMapStore

	//svc -> gw
	//route <-> gw
	//secret -> gw
}

func (g *GWLister) GetGatewayToRoutes(gwNsName string) []string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	var routes []string
	_, listeners := g.gatewayToListenerStore.Get(gwNsName)
	for _, listener := range listeners.([]string) {
		listenerSlice := strings.Split(listener, "/")

		found, route := g.gatewayListenerToRoute.Get(gwNsName + "/" + listenerSlice[0])
		if found {
			if !utils.HasElem(routes, route) {
				routes = append(routes, route.(string))
			}
		}
	}
	return routes
}

func (g *GWLister) UpdateGatewayRouteToHostname(ns, gw string, hostnames []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	key := getKeyForGateway(ns, gw)
	g.gatewayRouteToHostname.AddOrUpdate(key, hostnames)

}
func (g *GWLister) GetGatewayRouteToHostname(ns, gw string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	key := getKeyForGateway(ns, gw)
	found, hostnames := g.gatewayRouteToHostname.Get(key)
	if found {
		return true, hostnames.([]string)
	}
	return false, make([]string, 0)
}

func (g *GWLister) UpdateRouteToGateway(routeNsName string, gateways []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.routeToGateway.AddOrUpdate(routeNsName, gateways)

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

func (g *GWLister) IsGatewayInStore(gwNsName string) bool {
	gateways := g.gatewayToListenerStore.GetAllKeys()
	return utils.HasElem(gateways, gwNsName)
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

func (g *GWLister) UpdateGatewayToListener(gwNsName string, listeners []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayToListenerStore.AddOrUpdate(gwNsName, listeners)
}
func (g *GWLister) GetGatewayListenerToHostname(ns, gw, listner string) string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	key := getKeyForGateway(ns, gw) + "/" + listner
	_, obj := g.gatewayListenerToHostname.Get(key)

	return obj.(string)
}
func (g *GWLister) UpdateGatewayListenerToHostname(gwListenerNsName, hostname string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayListenerToHostname.AddOrUpdate(gwListenerNsName, hostname)
}

func (g *GWLister) GetGatewayToListeners(ns, gw string) []string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	key := getKeyForGateway(ns, gw)

	_, listenerList := g.gatewayToListenerStore.Get(key)
	return listenerList.([]string)

}

func (g *GWLister) DeleteGatewayToGatewayClass(ns, gw string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	key := getKeyForGateway(ns, gw)
	found, gwClass := g.gatewayToGatewayClassStore.Get(key)
	if found {
		g.gatewayToGatewayClassStore.Delete(key)
		g.gatewayToListenerStore.Delete(key)
		_, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass.(string))
		gatewayListObj := gatewayList.([]string)
		utils.Remove(gatewayListObj, key)
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass.(string), gatewayListObj)
	}
}

func getKeyForGateway(ns, gw string) string {
	return ns + "/" + gw
}

func (g *GWLister) DeleteGatewayListenerToRoute(gwNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayListenerToRoute.Delete(gwNsName)
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

func (g *GWLister) GetRouteToGatewayListener(routeType, routeNsName string) []string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	routeKey := routeType + "/" + routeNsName
	_, obj := g.routeToGatewayListener.Get(routeKey)
	return obj.([]string)

}
func (g *GWLister) UpdateGatewayListenerRouteMappings(gwListener, routeType, routeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	routeKey := routeType + "/" + routeNsName

	g.gatewayListenerToRoute.AddOrUpdate(gwListener, routeKey)

	// update route to gw mapping
	if found, obj := g.routeToGatewayListener.Get(routeKey); found {
		gwListeners := obj.([]string)
		if !utils.HasElem(gwListeners, gwListener) {
			gwListeners = append(gwListeners, gwListener)
		}
		g.routeToGatewayListener.AddOrUpdate(routeKey, gwListeners)
	} else {
		gwListeners := []string{gwListener}
		g.routeToGatewayListener.AddOrUpdate(routeKey, gwListeners)
	}
}

func (g *GWLister) DeleteGatewayListenerRouteMappings(routeType, routeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	routeKey := routeType + "/" + routeNsName
	if found, obj := g.routeToGatewayListener.Get(routeKey); found {
		gatewayListenerList := obj.([]string)
		for _, gatewayListener := range gatewayListenerList {
			g.gatewayListenerToRoute.Delete(gatewayListener)
		}
		g.routeToGatewayListener.Delete(routeKey)
	}

}

func (g *GWLister) GetServiceToGateway(serviceNsName string) ([]string, bool) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	found, obj := g.serviceToGateway.Get(serviceNsName)
	if found {
		return obj.([]string), true
	}
	return []string{}, false
}

func (g *GWLister) UpdateServiceToGateway(serviceNsName string, gatewayList []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.serviceToGateway.AddOrUpdate(serviceNsName, gatewayList)
}

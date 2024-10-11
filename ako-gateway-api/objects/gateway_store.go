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
	"fmt"
	"sync"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var gwLister *GWLister
var gwonce sync.Once

func GatewayApiLister() *GWLister {
	gwonce.Do(func() {
		gwLister = &GWLister{
			gatewayClassStore:              objects.NewObjectMapStore(),
			gatewayToGatewayClassStore:     objects.NewObjectMapStore(),
			gatewayClassToGatewayStore:     objects.NewObjectMapStore(),
			gatewayToListenerStore:         objects.NewObjectMapStore(),
			routeToGateway:                 objects.NewObjectMapStore(),
			routeToGatewayListener:         objects.NewObjectMapStore(),
			gatewayToRoute:                 objects.NewObjectMapStore(),
			serviceToGateway:               objects.NewObjectMapStore(),
			gatewayToService:               objects.NewObjectMapStore(),
			serviceToRoute:                 objects.NewObjectMapStore(),
			routeToService:                 objects.NewObjectMapStore(),
			secretToGateway:                objects.NewObjectMapStore(),
			gatewayToSecret:                objects.NewObjectMapStore(),
			routeToChildVS:                 objects.NewObjectMapStore(),
			gatewayToHostnameStore:         objects.NewObjectMapStore(),
			gatewayListenerToHostnameStore: objects.NewObjectMapStore(),
			gatewayRouteToHostnameStore:    objects.NewObjectMapStore(),
			gatewayRouteToHTTPSPGPoolStore: objects.NewObjectMapStore(),
			podToServiceStore:              objects.NewObjectMapStore(),
			gatewayToStatus:                objects.NewObjectMapStore(),
			routeToStatus:                  objects.NewObjectMapStore(),
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

	// routeType/routeNs/routeName -> [namespace/gateway, ...]
	routeToGateway *objects.ObjectMapStore

	routeToGatewayListener *objects.ObjectMapStore

	// namespace/gateway -> [routeType/routeNs/routeName, ...]
	gatewayToRoute *objects.ObjectMapStore

	// serviceNs/serviceName -> [namespace/gateway, ...]
	serviceToGateway *objects.ObjectMapStore

	// namespace/gateway -> [serviceNs/serviceName, ...]
	gatewayToService *objects.ObjectMapStore

	// serviceNs/serviceName -> [routeType/routeNs/routeName, ...]
	serviceToRoute *objects.ObjectMapStore

	// routeType/routeNs/routeName -> [serviceNs/serviceName, ...]
	routeToService *objects.ObjectMapStore

	// secretNs/secretName -> [namespace/gateway, ...]
	secretToGateway *objects.ObjectMapStore

	// namespace/gateway -> [secretNs/secretName, ...]
	gatewayToSecret *objects.ObjectMapStore

	// routeType/routeNs/routeName -> [childvs, ...]
	routeToChildVS *objects.ObjectMapStore

	//check overlap across gateways
	gatewayToHostnameStore *objects.ObjectMapStore

	//namespace/gateway/listener -> hostname
	gatewayListenerToHostnameStore *objects.ObjectMapStore

	//FQDNs in parent VS
	//gatewayns/gatewayname -> [hostname, ...]
	gatewayRouteToHostnameStore *objects.ObjectMapStore

	// HTTPPS, PG, Pool in parent VS
	// gatewayns/gatewayname + routenamespace/routename --> [HTTPPS, PG, Pool]
	gatewayRouteToHTTPSPGPoolStore *objects.ObjectMapStore

	//Pods -> Service Mapping for NPL
	//podNs/podName -> [svcNs/svcName, ...]
	podToServiceStore *objects.ObjectMapStore

	// namespace/gateway -> gateway Status
	gatewayToStatus *objects.ObjectMapStore

	// routeType/routeNs/routeName -> route Status
	routeToStatus *objects.ObjectMapStore
}

type GatewayRouteKind struct {
	Group string
	Kind  string
}

type GatewayListenerStore struct {
	Name              string
	Port              int32
	Protocol          string
	Gateway           string
	AllowedRouteNs    string
	AllowedRouteTypes []GatewayRouteKind
}

// This struct is used to store HTTPPS, PG, Pool associated with Parent VS (HTTPRoute that is mapped to parent VS)

type HTTPPSPGPool struct {
	HTTPPS    []string
	PoolGroup []string
	Pool      []string
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

func (g *GWLister) UpdateGatewayToGatewayClass(gwNsName, gwClass string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	//remove gateway from old class list
	if found, oldGwClass := g.gatewayToGatewayClassStore.Get(gwNsName); found {
		oldGwClassObj := oldGwClass.(string)
		if ok, gatewayList := g.gatewayClassToGatewayStore.Get(oldGwClassObj); ok {
			gatewayListObj := gatewayList.([]string)
			gatewayListObj = utils.Remove(gatewayListObj, gwNsName)
			if len(gatewayListObj) == 0 {
				g.gatewayClassToGatewayStore.Delete(oldGwClassObj)
			} else {
				g.gatewayClassToGatewayStore.AddOrUpdate(oldGwClassObj, gatewayListObj)
			}
		}
	}
	g.gatewayToGatewayClassStore.AddOrUpdate(gwNsName, gwClass)
	found, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	if !found {
		gatewayList = make([]string, 0)
	}
	gatewayListObj := gatewayList.([]string)
	if !utils.HasElem(gatewayListObj, gwNsName) {
		gatewayListObj = append(gatewayListObj, gwNsName)
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass, gatewayListObj)
	}
}

func (g *GWLister) UpdateGatewayToHostnames(gwNsName string, hostnames []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayToHostnameStore.AddOrUpdate(gwNsName, hostnames)
}

func (g *GWLister) UpdateGatewayToListener(gwNsName string, listeners []GatewayListenerStore) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayToListenerStore.AddOrUpdate(gwNsName, listeners)
}

func (g *GWLister) GetGatewayToListeners(gwNsName string) []GatewayListenerStore {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	_, listenerList := g.gatewayToListenerStore.Get(gwNsName)
	if listenerList != nil {
		return listenerList.([]GatewayListenerStore)
	}
	return nil
}

func (g *GWLister) UpdateGatewayToGatewayStatusMapping(gwName string, gwStatus *gatewayv1.GatewayStatus) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	g.gatewayToStatus.AddOrUpdate(gwName, gwStatus)
}

func (g *GWLister) DeleteGatewayToGatewayStatusMapping(gwName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	g.gatewayToStatus.Delete(gwName)
}

func (g *GWLister) GetGatewayToGatewayStatusMapping(gwName string) *gatewayv1.GatewayStatus {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()
	found, gatewayList := g.gatewayToStatus.Get(gwName)
	if !found {
		return nil
	}
	return gatewayList.(*gatewayv1.GatewayStatus)
}

func (g *GWLister) UpdateRouteToRouteStatusMapping(routeTypeNamespaceName string, routeStatus *gatewayv1.RouteStatus) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	g.routeToStatus.AddOrUpdate(routeTypeNamespaceName, routeStatus)
}

func (g *GWLister) DeleteRouteToRouteStatusMapping(routeTypeNamespaceName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	g.routeToStatus.Delete(routeTypeNamespaceName)
}

func (g *GWLister) GetRouteToRouteStatusMapping(routeTypeNamespaceName string) *gatewayv1.RouteStatus {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()
	found, routeList := g.routeToStatus.Get(routeTypeNamespaceName)
	if !found {
		return nil
	}
	return routeList.(*gatewayv1.RouteStatus)
}

//=====All route <-> gateway mappings go here.

func (g *GWLister) GetRouteToGateway(routeTypeNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.routeToGateway.Get(routeTypeNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) GetRouteToGatewayListener(routeTypeNsName string) []GatewayListenerStore {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.routeToGatewayListener.Get(routeTypeNsName); found {
		return obj.([]GatewayListenerStore)
	}
	return []GatewayListenerStore{}
}

func (g *GWLister) GetGatewayToRoute(gwNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.gatewayToRoute.Get(gwNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) UpdateGatewayRouteMappings(gwNsName string, gwListeners []GatewayListenerStore, routeTypeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// update route to gateway mapping
	if found, gwNsNameList := g.routeToGateway.Get(routeTypeNsName); found {
		gwNsNameListObj := gwNsNameList.([]string)
		if !utils.HasElem(gwNsNameListObj, gwNsName) {
			gwNsNameListObj = append(gwNsNameListObj, gwNsName)
			g.routeToGateway.AddOrUpdate(routeTypeNsName, gwNsNameListObj)
		}
	} else {
		g.routeToGateway.AddOrUpdate(routeTypeNsName, []string{gwNsName})
	}

	if found, gwListenerList := g.routeToGatewayListener.Get(routeTypeNsName); found {
		gwListenerListObj := gwListenerList.([]GatewayListenerStore)
		for _, gwListener := range gwListeners {
			if !utils.HasElem(gwListenerList, gwListener) {
				gwListenerListObj = append(gwListenerListObj, gwListener)
			}
		}
		g.routeToGatewayListener.AddOrUpdate(routeTypeNsName, gwListenerListObj)
	} else {
		g.routeToGatewayListener.AddOrUpdate(routeTypeNsName, gwListeners)
	}

	// update gateway to route mapping
	if found, routeTypeNsNameList := g.gatewayToRoute.Get(gwNsName); found {
		routeTypeNsNameListObj := routeTypeNsNameList.([]string)
		if !utils.HasElem(routeTypeNsNameListObj, routeTypeNsName) {
			routeTypeNsNameListObj = append(routeTypeNsNameListObj, routeTypeNsName)
			g.gatewayToRoute.AddOrUpdate(gwNsName, routeTypeNsNameListObj)
		}
	} else {
		g.gatewayToRoute.AddOrUpdate(gwNsName, []string{routeTypeNsName})
	}
}

//=====All gateway <-> service mappings go here.

func (g *GWLister) GetServiceToGateway(svcNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.serviceToGateway.Get(svcNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) UpdateGatewayServiceMappings(gwNsName, svcNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// update gateway to service mapping
	if found, svcNsNameList := g.gatewayToService.Get(gwNsName); found {
		svcNsNameListObj := svcNsNameList.([]string)
		if !utils.HasElem(svcNsNameListObj, svcNsName) {
			svcNsNameListObj = append(svcNsNameListObj, svcNsName)
			g.gatewayToService.AddOrUpdate(gwNsName, svcNsNameListObj)
		}
	} else {
		g.gatewayToService.AddOrUpdate(gwNsName, []string{svcNsName})
	}
	// update service to gateway mapping
	if found, gwNsNameList := g.serviceToGateway.Get(gwNsName); found {
		gwNsNameListObj := gwNsNameList.([]string)
		if !utils.HasElem(gwNsNameListObj, gwNsName) {
			gwNsNameListObj = append(gwNsNameListObj, gwNsName)
			g.serviceToGateway.AddOrUpdate(svcNsName, gwNsNameListObj)
		}
	} else {
		g.serviceToGateway.AddOrUpdate(svcNsName, []string{gwNsName})
	}
}

func (g *GWLister) DeleteGatewayServiceMappings(gwNsName string, svcNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	var gwRoutesObj []string
	found, gwRoutes := g.gatewayToRoute.Get(gwNsName)
	if found {
		gwRoutesObj = gwRoutes.([]string)
	}
	var svcRoutesObj []string
	found, svcRoutes := g.serviceToRoute.Get(svcNsName)
	if found {
		svcRoutesObj = svcRoutes.([]string)
	}
	matchCount := 0
	for _, gwRoute := range gwRoutesObj {
		if utils.HasElem(svcRoutesObj, gwRoute) {
			matchCount++
		}
	}
	if matchCount == 0 {
		// delete gateway to service mapping
		if found, svcNsNameList := g.gatewayToService.Get(gwNsName); found {
			svcNsNameListObj := svcNsNameList.([]string)
			svcNsNameListObj = utils.Remove(svcNsNameListObj, svcNsName)
			if len(svcNsNameListObj) == 0 {
				g.gatewayToService.Delete(gwNsName)
			} else {
				g.gatewayToService.AddOrUpdate(gwNsName, svcNsNameListObj)
			}
		}

		// delete service to gateway mapping
		if found, gwNsNameList := g.serviceToGateway.Get(svcNsName); found {
			gwNsNameListObj := gwNsNameList.([]string)
			gwNsNameListObj = utils.Remove(gwNsNameListObj, gwNsName)
			if len(gwNsNameListObj) == 0 {
				g.serviceToGateway.Delete(svcNsName)
			} else {
				g.serviceToGateway.AddOrUpdate(svcNsName, gwNsNameListObj)
			}
		}
	}
}

//=====All route <-> service mappings go here.

func (g *GWLister) GetRouteToService(routeTypeNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.routeToService.Get(routeTypeNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) GetServiceToRoute(svcNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.serviceToRoute.Get(svcNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) UpdateRouteServiceMappings(routeTypeNsName, svcNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// update route to service mapping
	if found, svcNsNameList := g.routeToService.Get(routeTypeNsName); found {
		svcNsNameListObj := svcNsNameList.([]string)
		if !utils.HasElem(svcNsNameListObj, svcNsName) {
			svcNsNameListObj = append(svcNsNameListObj, svcNsName)
			g.routeToService.AddOrUpdate(routeTypeNsName, svcNsNameListObj)
		}
	} else {
		g.routeToService.AddOrUpdate(routeTypeNsName, []string{svcNsName})
	}

	// update service to route mapping
	if found, routeTypeNsNameList := g.serviceToRoute.Get(svcNsName); found {
		routeTypeNsNameListObj := routeTypeNsNameList.([]string)
		if !utils.HasElem(routeTypeNsNameListObj, routeTypeNsName) {
			routeTypeNsNameListObj = append(routeTypeNsNameListObj, routeTypeNsName)
			g.serviceToRoute.AddOrUpdate(svcNsName, routeTypeNsNameListObj)
		}
	} else {
		g.serviceToRoute.AddOrUpdate(svcNsName, []string{routeTypeNsName})
	}
}

func (g *GWLister) DeleteRouteToServiceMappings(routeTypeNsName, svcNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// delete service to route mapping
	if found, routeTypeNsNameList := g.serviceToRoute.Get(svcNsName); found {
		routeTypeNsNameListObj := routeTypeNsNameList.([]string)
		routeTypeNsNameListObj = utils.Remove(routeTypeNsNameListObj, routeTypeNsName)
		if len(routeTypeNsNameListObj) == 0 {
			g.serviceToRoute.Delete(svcNsName)
		} else {
			g.serviceToRoute.AddOrUpdate(svcNsName, routeTypeNsNameListObj)
		}
	}

	// delete route to service mapping
	if found, svcNsNameList := g.routeToService.Get(routeTypeNsName); found {
		svcNsNameListObj := svcNsNameList.([]string)
		svcNsNameListObj = utils.Remove(svcNsNameListObj, svcNsName)
		if len(svcNsNameListObj) == 0 {
			g.routeToService.Delete(routeTypeNsName)
		} else {
			g.routeToService.AddOrUpdate(routeTypeNsName, svcNsNameListObj)
		}
	}
}

func (g *GWLister) DeleteRouteToGatewayMappings(routeTypeNsName, gwNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// delete gateway to route mapping
	if found, routeTypeNsNameList := g.gatewayToRoute.Get(gwNsName); found {
		routeTypeNsNameListObj := routeTypeNsNameList.([]string)
		routeTypeNsNameListObj = utils.Remove(routeTypeNsNameListObj, routeTypeNsName)
		if len(routeTypeNsNameListObj) == 0 {
			g.gatewayToRoute.Delete(gwNsName)
		} else {
			g.gatewayToRoute.AddOrUpdate(gwNsName, routeTypeNsNameListObj)
		}
	}

	// delete route to gateway mapping
	if found, gwNsNameList := g.routeToGateway.Get(routeTypeNsName); found {
		gwNsNameListObj := gwNsNameList.([]string)
		gwNsNameListObj = utils.Remove(gwNsNameListObj, gwNsName)
		if len(gwNsNameListObj) == 0 {
			g.routeToGateway.Delete(routeTypeNsName)
		} else {
			g.routeToGateway.AddOrUpdate(routeTypeNsName, gwNsNameListObj)
		}
	}
}

//=====All Gateway <-> Secret go here

func (g *GWLister) GetSecretToGateway(secretNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()
	if found, obj := g.secretToGateway.Get(secretNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}

}
func (g *GWLister) UpdateGatewayToSecret(gwNsName string, secretNsNameList []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	var removedSecrets []string
	//update list of secrets store
	found, obj := g.gatewayToSecret.Get(gwNsName)
	if found {
		//find removed secrets
		for _, secret := range obj.([]string) {
			secretRemoved := true
			for _, newSecret := range secretNsNameList {
				if secret == newSecret {
					secretRemoved = false
					break
				}
			}
			if secretRemoved {
				removedSecrets = append(removedSecrets, secret)
			}
		}
	}
	//update new secrets to gateway
	g.gatewayToSecret.AddOrUpdate(gwNsName, secretNsNameList)

	//delete secret to gateway mapping for removed secret
	for _, secret := range removedSecrets {
		if found, gwNsNameList := g.secretToGateway.Get(secret); found {
			gwNsNameListObj := gwNsNameList.([]string)
			gwNsNameListObj = utils.Remove(gwNsNameListObj, gwNsName)
			if len(gwNsNameListObj) == 0 {
				g.secretToGateway.Delete(secret)
			} else {
				g.secretToGateway.AddOrUpdate(secret, gwNsNameListObj)
			}
		}
	}

	//add secret to gateway mapping for new secrets
	for _, secret := range secretNsNameList {
		if found, gwNsNameList := g.secretToGateway.Get(secret); found {
			gwNsNameListObj := gwNsNameList.([]string)
			if !utils.HasElem(gwNsNameListObj, gwNsName) {
				gwNsNameListObj = append(gwNsNameListObj, gwNsName)
				g.secretToGateway.AddOrUpdate(secret, gwNsNameListObj)
			}
		} else {
			g.secretToGateway.AddOrUpdate(secret, []string{gwNsName})
		}
	}
}

func (g *GWLister) DeleteGatewayFromStore(gwNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// delete gateway to secrets

	if found, secretNsNameList := g.gatewayToSecret.Get(gwNsName); found {
		for _, secretNsName := range secretNsNameList.([]string) {
			if found, gwNsNameList := g.serviceToGateway.Get(secretNsName); found {
				gwNsNameListObj := gwNsNameList.([]string)
				gwNsNameListObj = utils.Remove(gwNsNameListObj, gwNsName)
				if len(gwNsNameListObj) == 0 {
					g.secretToGateway.Delete(secretNsName)
				} else {
					g.secretToGateway.AddOrUpdate(secretNsName, gwNsNameListObj)
				}
			}
		}
		g.gatewayToSecret.Delete(gwNsName)
	}

	// delete gateway to service
	if found, serviceNsNameList := g.gatewayToService.Get(gwNsName); found {
		for _, serviceNsName := range serviceNsNameList.([]string) {
			if found, gwNsNameList := g.serviceToGateway.Get(serviceNsName); found {
				gwNsNameListObj := gwNsNameList.([]string)
				gwNsNameListObj = utils.Remove(gwNsNameListObj, gwNsName)
				if len(gwNsNameListObj) == 0 {
					g.serviceToGateway.Delete(serviceNsName)
				} else {
					g.serviceToGateway.AddOrUpdate(serviceNsName, gwNsNameListObj)
				}
			}
		}
		g.gatewayToService.Delete(gwNsName)
	}

	// delete gateway to route
	if found, routeNsNameList := g.gatewayToRoute.Get(gwNsName); found {
		for _, routeNsName := range routeNsNameList.([]string) {
			if found, gwNsNameList := g.routeToGateway.Get(routeNsName); found {
				gwNsNameListObj := gwNsNameList.([]string)
				gwNsNameListObj = utils.Remove(gwNsNameListObj, gwNsName)
				if len(gwNsNameListObj) == 0 {
					g.routeToGateway.Delete(routeNsName)
				} else {
					g.routeToGateway.AddOrUpdate(routeNsName, gwNsNameListObj)
				}
			}
		}
		g.gatewayToRoute.Delete(gwNsName)
	}

	// delete gateway to gatewayclass
	if found, gatewayClass := g.gatewayToGatewayClassStore.Get(gwNsName); found {
		gatewayClassObj := gatewayClass.(string)
		if found, gwNsNameList := g.gatewayClassToGatewayStore.Get(gatewayClassObj); found {
			gwNsNameListObj := gwNsNameList.([]string)
			gwNsNameListObj = utils.Remove(gwNsNameListObj, gwNsName)
			if len(gwNsNameListObj) == 0 {
				g.gatewayClassToGatewayStore.Delete(gatewayClassObj)
			} else {
				g.gatewayClassToGatewayStore.AddOrUpdate(gatewayClassObj, gwNsNameListObj)
			}
		}
		g.gatewayToGatewayClassStore.Delete(gwNsName)
	}

}

// =====All route <-> child vs go here.
func (g *GWLister) GetRouteToChildVS(routeTypeNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, obj := g.routeToChildVS.Get(routeTypeNsName); found {
		return true, obj.([]string)
	}
	return false, []string{}
}

func (g *GWLister) DeleteRouteToChildVS(routeTypeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.routeToChildVS.Delete(routeTypeNsName)
}

func (g *GWLister) UpdateRouteChildVSMappings(routeTypeNsName, childVS string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// update route to child vs mapping
	if found, childVSList := g.routeToChildVS.Get(routeTypeNsName); found {
		childVSListObj := childVSList.([]string)
		if !utils.HasElem(childVSListObj, childVS) {
			childVSListObj = append(childVSListObj, childVS)
			g.routeToChildVS.AddOrUpdate(routeTypeNsName, childVSListObj)
		}
	} else {
		g.routeToChildVS.AddOrUpdate(routeTypeNsName, []string{childVS})
	}
}

func (g *GWLister) DeleteRouteChildVSMappings(routeTypeNsName, childVS string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	// delete route to child vs mapping
	if found, childVSList := g.routeToChildVS.Get(routeTypeNsName); found {
		childVSListObj := childVSList.([]string)
		childVSListObj = utils.Remove(childVSListObj, childVS)
		if len(childVSListObj) == 0 {
			g.routeToChildVS.Delete(routeTypeNsName)
		} else {
			g.routeToChildVS.AddOrUpdate(routeTypeNsName, childVSListObj)
		}
	}
}

//=====All route functions go here.

func (g *GWLister) DeleteRouteFromStore(routeTypeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	//delete route to gateway listener
	g.routeToGatewayListener.Delete(routeTypeNsName)

	//delete route to gateway
	found, gatewayList := g.routeToGateway.Get(routeTypeNsName)
	var gatewayListObj, svcListObj []string
	if found {
		gatewayListObj = gatewayList.([]string)
	}
	for _, gwNsName := range gatewayListObj {
		if found, routeNsNameList := g.gatewayToRoute.Get(gwNsName); found {
			routeNsNameListObj := routeNsNameList.([]string)
			routeNsNameListObj = utils.Remove(routeNsNameListObj, routeTypeNsName)
			if len(routeNsNameListObj) == 0 {
				g.gatewayToRoute.Delete(gwNsName)
			} else {
				g.gatewayToRoute.AddOrUpdate(gwNsName, routeNsNameListObj)
			}
			// remove hostname mapping
			gwRouteNsName := fmt.Sprintf("%s/%s", gwNsName, routeTypeNsName)
			g.gatewayRouteToHostnameStore.Delete(gwRouteNsName)
		}
	}
	g.routeToGateway.Delete(routeTypeNsName)

	//delete route to service
	found, svcList := g.routeToService.Get(routeTypeNsName)
	if found {
		svcListObj = svcList.([]string)
	}
	for _, svcNsName := range svcListObj {
		if found, routeNsNameList := g.serviceToRoute.Get(svcNsName); found {
			routeNsNameListObj := routeNsNameList.([]string)
			routeNsNameListObj = utils.Remove(routeNsNameListObj, routeTypeNsName)
			if len(routeNsNameListObj) == 0 {
				g.serviceToRoute.Delete(svcNsName)
			} else {
				g.serviceToRoute.AddOrUpdate(svcNsName, routeNsNameListObj)
			}
		}
	}
	g.routeToService.Delete(routeTypeNsName)

	//update gateway to service mappings after route deletion
	for _, gwNsName := range gatewayListObj {
		var gwRoutesObj []string
		found, gwRoutes := g.gatewayToRoute.Get(gwNsName)
		if found {
			gwRoutesObj = gwRoutes.([]string)
		}
		for _, svcNsName := range svcListObj {
			var svcRoutesObj []string
			found, svcRoutes := g.serviceToRoute.Get(svcNsName)
			if found {
				svcRoutesObj = svcRoutes.([]string)
			}
			matchCount := 0
			for _, gwRoute := range gwRoutesObj {
				if utils.HasElem(svcRoutesObj, gwRoute) {
					matchCount++
				}
			}
			if matchCount == 0 {
				//no routes to map gateway and service, remove mapping from store as well
				if found, gwSvcList := g.gatewayToService.Get(gwNsName); found {
					gwSvcListObj := gwSvcList.([]string)
					gwSvcListObj = utils.Remove(gwSvcListObj, svcNsName)
					if len(gwSvcListObj) == 0 {
						g.gatewayToService.Delete(gwNsName)
					} else {
						g.gatewayToService.AddOrUpdate(gwNsName, gwSvcListObj)
					}
				}
				if found, svcGwList := g.serviceToGateway.Get(gwNsName); found {
					svcGwListObj := svcGwList.([]string)
					svcGwListObj = utils.Remove(svcGwListObj, gwNsName)
					if len(svcGwListObj) == 0 {
						g.serviceToGateway.Delete(gwNsName)
					} else {
						g.serviceToGateway.AddOrUpdate(svcNsName, svcGwListObj)
					}
				}
			}
		}
	}
}

func (g *GWLister) DeleteRouteServiceMappings(routeTypeNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	if found, svcNsNameList := g.routeToService.Get(routeTypeNsName); found {
		for _, svcNsName := range svcNsNameList.([]string) {
			if found, routeTypeNsNameList := g.serviceToRoute.Get(svcNsName); found {
				routeTypeNsNameListObj := routeTypeNsNameList.([]string)
				routeTypeNsNameListObj = utils.Remove(routeTypeNsNameListObj, routeTypeNsName)
				if len(routeTypeNsNameListObj) == 0 {
					g.serviceToRoute.Delete(svcNsName)
				} else {
					g.serviceToRoute.AddOrUpdate(svcNsName, routeTypeNsNameListObj)
				}
			}
		}

		if found, gwNsNameList := g.routeToGateway.Get(routeTypeNsName); found {
			for _, gwNsName := range gwNsNameList.([]string) {
				if found, gwSvcNsNameList := g.gatewayToService.Get(gwNsName); found {
					gwSvcNsNameListObj := gwSvcNsNameList.([]string)
					for _, svcNsName := range svcNsNameList.([]string) {
						gwSvcNsNameListObj = utils.Remove(gwSvcNsNameListObj, svcNsName)
					}
					if len(gwSvcNsNameListObj) == 0 {
						g.gatewayToService.Delete(gwNsName)
					} else {
						g.gatewayToService.AddOrUpdate(gwNsName, gwSvcNsNameListObj)
					}
				}
			}
		}
		g.routeToService.Delete(routeTypeNsName)
	}
}

//=====All gateway/route <-> hostname go here.

func (g *GWLister) GetGatewayListenerToHostname(gwListenerNsName string) string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	_, obj := g.gatewayListenerToHostnameStore.Get(gwListenerNsName)
	return obj.(string)
}
func (g *GWLister) UpdateGatewayListenerToHostname(gwListenerNsName, hostname string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayListenerToHostnameStore.AddOrUpdate(gwListenerNsName, hostname)
}

func (g *GWLister) UpdateGatewayRouteToHostname(gwRouteNsName string, hostnames []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.gatewayRouteToHostnameStore.AddOrUpdate(gwRouteNsName, hostnames)

}

func (g *GWLister) GetGatewayRouteToHostname(gwRouteNsName string) (bool, []string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	found, hostnames := g.gatewayRouteToHostnameStore.Get(gwRouteNsName)
	if found {
		return true, hostnames.([]string)
	}
	return false, []string{}
}

// == All GW+route to HTTPS, PG , pool mapping
func (g *GWLister) UpdateGatewayRouteToHTTPPSPGPool(gwRouteNsName string, httpPSPGPool HTTPPSPGPool) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	var localHTTPPSPGPool HTTPPSPGPool
	localHTTPPSPGPool.HTTPPS = append(localHTTPPSPGPool.HTTPPS, httpPSPGPool.HTTPPS...)
	localHTTPPSPGPool.PoolGroup = append(localHTTPPSPGPool.PoolGroup, httpPSPGPool.PoolGroup...)
	localHTTPPSPGPool.Pool = append(localHTTPPSPGPool.Pool, httpPSPGPool.Pool...)
	g.gatewayRouteToHTTPSPGPoolStore.AddOrUpdate(gwRouteNsName, localHTTPPSPGPool)

}

func (g *GWLister) GetGatewayRouteToHTTPSPGPool(gwRouteNsName string) (bool, HTTPPSPGPool) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	found, hostnames := g.gatewayRouteToHTTPSPGPoolStore.Get(gwRouteNsName)
	if found {
		return true, hostnames.(HTTPPSPGPool)
	}
	return false, HTTPPSPGPool{}
}

func (g *GWLister) DeleteGatewayRouteToHTTPSPGPool(gwRouteNsName string) {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()
	g.gatewayRouteToHTTPSPGPoolStore.Delete(gwRouteNsName)
}

//Pods <-> Service

func (g *GWLister) GetPodsToService(podNsName string) []string {
	g.gwLock.RLock()
	defer g.gwLock.RUnlock()

	if found, services := g.podToServiceStore.Get(podNsName); found {
		return services.([]string)
	}
	return []string{}
}

func (g *GWLister) UpdatePodsToService(podNsName string, services []string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.podToServiceStore.AddOrUpdate(podNsName, services)

}
func (g *GWLister) DeletePodsToService(podNsName string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	g.podToServiceStore.Delete(podNsName)
}

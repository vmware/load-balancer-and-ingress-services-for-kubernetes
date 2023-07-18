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
}

func (g *GWLister) IsGatewayClassPresent(gwClass string) bool {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	found, _ := g.gatewayClassStore.Get(gwClass)
	return found
}

func (g *GWLister) UpdateGatewayClass(gwClass string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	found, _ := g.gatewayClassStore.Get(gwClass)
	if !found {
		g.gatewayClassStore.AddOrUpdate(gwClass, struct{}{})
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass, make([]string, 0))
	}
}

func (g *GWLister) DeleteGatewayClass(gwClass string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	found, _ := g.gatewayClassStore.Get(gwClass)
	if found {
		g.deleteGatewayClassToGateway(gwClass)
		g.gatewayClassStore.Delete(gwClass)
	}
}

func (g *GWLister) GetGatewayClassToGateway(gwClass string) []string {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	_, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	return gatewayList.([]string)
}

// do not use, instead use UpdateGatewayToGatewayClass
func (g *GWLister) updateGatewayClassToGateway(gwClass, ns, gw string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	_, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	gatewayListObj := gatewayList.([]string)
	gwObj := getKeyForGateway(ns, gw)

	found := utils.HasElem(gatewayListObj, gwObj)
	if !found {
		gatewayListObj = append(gatewayListObj, gwObj)
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass, gatewayListObj)

	}
}

// do not use, instead use DeleteGatewayToGatewayClass
func (g *GWLister) removeGatewayClassToGateway(gwClass, ns, gw string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	_, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	gatewayListObj := gatewayList.([]string)
	gwObj := getKeyForGateway(ns, gw)

	found := utils.HasElem(gatewayListObj, gwObj)
	if found {
		utils.Remove(gatewayListObj, gwObj)
		g.gatewayClassToGatewayStore.AddOrUpdate(gwClass, gatewayListObj)
	}
}

func (g *GWLister) deleteGatewayClassToGateway(gwClass string) {

	_, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	gatewayListObj := gatewayList.([]string)
	for _, key := range gatewayListObj {
		//DeleteGatewayToGatewayClass

		found, _ := g.gatewayToGatewayClassStore.Get(key)
		if found {
			g.gatewayToGatewayClassStore.Delete(key)
		}
	}
	g.gatewayClassToGatewayStore.Delete(gwClass)
}

func (g *GWLister) GetGatewayToGatewayClass(ns, gw string) string {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

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
	_, gatewayList := g.gatewayClassToGatewayStore.Get(gwClass)
	gatewayListObj := gatewayList.([]string)
	if !utils.HasElem(gatewayListObj, gw) {
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

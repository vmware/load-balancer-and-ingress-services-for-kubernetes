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
)

var gwLister *GWLister
var gwonce sync.Once

func GatewayApiLister() *GWLister {
	gwonce.Do(func() {
		gwLister = &GWLister{
			GatewayClassStore:     objects.NewObjectMapStore(),
			GatewayToGatewayClass: objects.NewObjectMapStore(),
			GatewayClassToGateway: objects.NewObjectMapStore(),
		}
	})
	return gwLister
}

type GWLister struct {
	gwLock sync.RWMutex

	//Gateways with AKO as controller
	GatewayClassStore *objects.ObjectMapStore

	//Gateway -> GatewayClass
	GatewayToGatewayClass *objects.ObjectMapStore

	//GatewayClass -> [gateway1, gateway2]
	GatewayClassToGateway *objects.ObjectMapStore
}

func (g *GWLister) IsGatewayClassPresent(gwClass string) bool {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	found, _ := g.GatewayClassStore.Get(gwClass)
	return found
}

func (g *GWLister) UpdateGatewayClass(gwClass string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	found, _ := g.GatewayClassStore.Get(gwClass)
	if !found {
		g.GatewayClassStore.AddOrUpdate(gwClass, struct{}{})
		g.GatewayClassToGateway.AddOrUpdate(gwClass, make([]string, 0))
	}
}

func (g *GWLister) DeleteGatewayClass(gwClass string) {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()
	found, _ := g.GatewayClassStore.Get(gwClass)
	if found {
		g.GatewayClassStore.Delete(gwClass)
		//TODO update GatewayToGatewayClass and GatewayClassToGateway
	}
}

func (g *GWLister) GetGatewayClassToGateway(gwClass string) []string {
	g.gwLock.Lock()
	defer g.gwLock.Unlock()

	_, gatewayList := g.GatewayClassToGateway.Get(gwClass)
	return gatewayList.([]string)
}

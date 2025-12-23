/*
 * Copyright 2019-2020 VMware, Inc.
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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var gwsvclister *SvcGWLister
var gwonce sync.Once

// This file builds cache relations for all services API objects.
// Relationships stored are: gatewayclass to gateway, service to gateway,
// GatewayClass is a cluster scoped resource.

func ServiceGWLister() *SvcGWLister {
	gwonce.Do(func() {
		gwsvclister = &SvcGWLister{
			GwClassGWStore:   NewObjectMapStore(),
			GwGwClassStore:   NewObjectMapStore(),
			GwListenersStore: NewObjectMapStore(),
			SvcGWStore:       NewObjectMapStore(),
			GwSvcsStore:      NewObjectMapStore(),
		}
	})
	return gwsvclister
}

type SvcGWLister struct {
	SvcGWLock sync.RWMutex

	// gwclass -> [ns1/gw1, ns1/gw2, ns2/gw3]
	GwClassGWStore *ObjectMapStore

	// nsX/gw1 -> gwclass
	GwGwClassStore *ObjectMapStore

	// the protocol and port mapped here are of the gateway listener config
	// that has the appropriate labels
	// nsX/gw1 -> [proto1/port1, proto2/port2]
	GwListenersStore *ObjectMapStore

	// ns1/svc -> nsX/gw1
	SvcGWStore *ObjectMapStore

	// the protocol and port mapped here are of the service
	// nsX/gw1 -> {proto1/port1: ns1/svc1, proto2/port2: ns2/svc2, ...}
	GwSvcsStore *ObjectMapStore
}

// Gateway <-> GatewayClass
func (v *SvcGWLister) GetGWclassToGateways(gwclass string) (bool, []string) {
	found, gatewayList := v.GwClassGWStore.Get(gwclass)
	if !found {
		return false, make([]string, 0)
	}
	return true, gatewayList.([]string)
}

func (v *SvcGWLister) GetGatewayToGWclass(gateway string) (bool, string) {
	found, gwClass := v.GwGwClassStore.Get(gateway)
	if !found {
		return false, ""
	}
	return true, gwClass.(string)
}

func (v *SvcGWLister) UpdateGatewayGWclassMappings(gateway, gwclass string) {
	v.SvcGWLock.Lock()
	defer v.SvcGWLock.Unlock()
	_, gatewayList := v.GetGWclassToGateways(gwclass)
	if !utils.HasElem(gatewayList, gateway) {
		gatewayList = append(gatewayList, gateway)
	}
	v.GwClassGWStore.AddOrUpdate(gwclass, gatewayList)
	v.GwGwClassStore.AddOrUpdate(gateway, gwclass)
}

func (v *SvcGWLister) RemoveGatewayGWclassMappings(gateway string) bool {
	v.SvcGWLock.Lock()
	defer v.SvcGWLock.Unlock()
	found, gwclass := v.GetGatewayToGWclass(gateway)
	if !found {
		return false
	}

	if found, gatewayList := v.GetGWclassToGateways(gwclass); found && utils.HasElem(gatewayList, gateway) {
		gatewayList = utils.Remove(gatewayList, gateway)
		if len(gatewayList) == 0 {
			v.GwClassGWStore.Delete(gwclass)
			return true
		}
		v.GwClassGWStore.AddOrUpdate(gwclass, gatewayList)
		return true
	}
	return false
}

// Gateway <-> Listeners
func (v *SvcGWLister) GetGWListeners(gateway string) (bool, []string) {
	found, listeners := v.GwListenersStore.Get(gateway)
	if !found {
		return false, make([]string, 0)
	}
	return true, listeners.([]string)
}

func (v *SvcGWLister) UpdateGWListeners(gateway string, listeners []string) {
	v.GwListenersStore.AddOrUpdate(gateway, listeners)
}

func (v *SvcGWLister) DeleteGWListeners(gateway string) bool {
	success := v.GwListenersStore.Delete(gateway)
	return success
}

//=====All service <-> gateway mappings go here. The mappings are updated only if the Gateway validation is successful in layer 1.

func (v *SvcGWLister) GetSvcToGw(service string) (bool, string) {
	found, gateway := v.SvcGWStore.Get(service)
	if !found {
		return false, ""
	}
	return true, gateway.(string)
}

func (v *SvcGWLister) GetGwToSvcs(gateway string) (bool, map[string][]string) {
	found, services := v.GwSvcsStore.Get(gateway)
	if !found {
		return false, make(map[string][]string)
	}
	return true, services.(map[string][]string)
}

func (v *SvcGWLister) UpdateGatewayMappings(gateway string, svcListener map[string][]string, service string) {
	v.SvcGWLock.Lock()
	defer v.SvcGWLock.Unlock()
	v.GwSvcsStore.AddOrUpdate(gateway, svcListener)
	v.SvcGWStore.AddOrUpdate(service, gateway)
}

func (v *SvcGWLister) RemoveGatewayMappings(gateway, service string) bool {
	v.SvcGWLock.Lock()
	defer v.SvcGWLock.Unlock()
	_, svcListeners := v.GetGwToSvcs(gateway)
	for portproto, svcs := range svcListeners {
		if utils.HasElem(svcs, service) {
			svcs = utils.Remove(svcs, service)
			if len(svcs) == 0 {
				delete(svcListeners, portproto)
				continue
			}
			svcListeners[portproto] = svcs
			v.GwSvcsStore.AddOrUpdate(gateway, svcListeners)
		}
	}

	if len(svcListeners) == 0 {
		if success := v.GwSvcsStore.Delete(gateway); !success {
			return false
		}
	}
	return v.SvcGWStore.Delete(service)
}

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

	"github.com/avinetworks/ako/pkg/utils"
)

var gwsvclister *SvcGWLister
var gwonce sync.Once

// This file builds cache relations for all services API objects.
// Relationships stored are: gatewayclass to gateway, service to gateway,
// GatewayClass is a cluster scoped resource.

func ServiceGWLister() *SvcGWLister {
	gwonce.Do(func() {
		gwsvclister = &SvcGWLister{
			gwSvcStore: NewObjectStore(),
		}
	})
	return gwsvclister
}

type SvcGWLister struct {
	gwSvcStore *ObjectStore
}

type SvcGWNSCache struct {
	namespace  string
	svcGWStore *ObjectMapStore
	svcGWLock  sync.RWMutex
}

func (v *SvcGWLister) SvcGWMappings(ns string) *SvcGWNSCache {
	return &SvcGWNSCache{
		namespace:  ns,
		svcGWStore: v.gwSvcStore.GetNSStore(ns),
	}
}

//=====All service to gateway mappings go here. The mappings are updated only if the Gateway validation is successful in layer 1.
// TODO : Do we have to map the services to listeners of the gateway instead of the gateway as as a whole?

func (v *SvcGWNSCache) GetSvcToGw(svcName string) (bool, []string) {
	// Need checks if it's found or not?
	found, gwNames := v.svcGWStore.Get(svcName)
	if !found {
		return false, make([]string, 0)
	}
	return true, gwNames.([]string)
}

func (v *SvcGWNSCache) DeleteSvcToGwMapping(svcName string) bool {
	// Need checks if it's found or not?
	success := v.svcGWStore.Delete(svcName)
	utils.AviLog.Debugf("Deleted the gateway mappings for svc: %s", svcName)
	return success
}

func (v *SvcGWNSCache) UpdateSvcToGwMapping(svcName string, gatewayList []string) {
	utils.AviLog.Debugf("Updated the mappings with svc: %s, gateways: %s", svcName, gatewayList)
	v.svcGWStore.AddOrUpdate(svcName, gatewayList)
}

/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

//This package gives relationship APIs to manage a multi-cluster ingress object.

var mciSvcListerInstance *SvcLister
var once sync.Once

func SharedMultiClusterIngressSvcLister() *SvcLister {
	once.Do(func() {
		mciSvcListerInstance = &SvcLister{
			svcIngStore:         NewObjectStore(),
			ingSvcStore:         NewObjectStore(),
			secretIngStore:      NewObjectStore(),
			ingSecretStore:      NewObjectStore(),
			secretHostNameStore: NewObjectStore(),
			ingHostStore:        NewObjectStore(),
			classIngStore:       NewObjectStore(),
			ingClassStore:       NewObjectStore(),
			svcSIStore:          NewObjectStore(),
			SISvcStore:          NewObjectStore(),
		}
	})
	return mciSvcListerInstance
}

type mciNSCache struct {
	svcSICache *ObjectMapStore
	SISvcCache *ObjectMapStore
	sync.RWMutex
}

func (v *SvcLister) MultiClusterIngressMappings(ns string) *SvcNSCache {
	svcNSCache := &SvcNSCache{
		namespace:       ns,
		svcIngObject:    v.svcIngStore.GetNSStore(ns),
		secretIngObject: v.secretIngStore.GetNSStore(ns),
		classIngObject:  v.classIngStore.GetNSStore(ns),
		IngNSCache: IngNSCache{
			ingSvcObjects: v.ingSvcStore.GetNSStore(ns),
		},
		SecretHostNameNSCache: SecretHostNameNSCache{
			secretHostNameObjects: v.secretHostNameStore.GetNSStore(ns),
		},
		SecretIngNSCache: SecretIngNSCache{
			ingSecretObjects: v.ingSecretStore.GetNSStore(ns),
		},
		IngHostCache: IngHostCache{
			ingHostObjects: v.ingHostStore.GetNSStore(ns),
		},
		IngClassNSCache: IngClassNSCache{
			ingClassObjects: v.ingClassStore.GetNSStore(ns),
		},
		mciNSCache: mciNSCache{
			svcSICache: v.svcSIStore.GetNSStore(ns),
			SISvcCache: v.SISvcStore.GetNSStore(ns),
		},
	}
	return svcNSCache
}

//=====All service to service imports mapping methods are here.

func (c *mciNSCache) GetSvcToSI(svcName string) (bool, []string) {
	c.Lock()
	defer c.Unlock()
	found, siNames := c.svcSICache.Get(svcName)
	if !found {
		return false, make([]string, 0)
	}
	return true, siNames.([]string)
}

func (c *mciNSCache) DeleteSvcToSIMapping(svcName string) bool {
	c.Lock()
	defer c.Unlock()
	success := c.svcSICache.Delete(svcName)
	utils.AviLog.Debugf("Deleted the service to service imports mappings for service with name: %s", svcName)
	return success
}

func (c *mciNSCache) UpdateSvcToSIMapping(svcName string, siList []string) {
	c.Lock()
	defer c.Unlock()
	utils.AviLog.Debugf("Updated the service to service imports mappings for service with name: %s, service imports: %s", svcName, siList)
	c.svcSICache.AddOrUpdate(svcName, siList)
}

//=====All service imports to service mapping methods are here.

func (c *mciNSCache) GetSIToSvc(siName string) (bool, []string) {
	c.Lock()
	defer c.Unlock()
	found, svcNames := c.SISvcCache.Get(siName)
	if !found {
		return false, make([]string, 0)
	}
	return true, svcNames.([]string)
}

func (c *mciNSCache) DeleteSIToSvcMapping(siName string) bool {
	c.Lock()
	defer c.Unlock()
	success := c.SISvcCache.Delete(siName)
	utils.AviLog.Debugf("Deleted the service to service imports mappings for service import with name: %s", siName)
	return success
}

func (c *mciNSCache) UpdateSIToSvcMapping(siName string, svcList []string) {
	c.Lock()
	defer c.Unlock()
	utils.AviLog.Debugf("Updated the service imports to service mappings for service import with name: %s, services: %s", siName, svcList)
	c.SISvcCache.AddOrUpdate(siName, svcList)
}

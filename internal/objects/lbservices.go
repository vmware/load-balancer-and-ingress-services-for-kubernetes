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

var lbinstance *lbLister
var lbonce sync.Once

func SharedlbLister() *lbLister {
	lbonce.Do(func() {
		lbinstance = &lbLister{
			lbStore:                     NewObjectMapStore(),
			sharedVipKeyToServicesStore: NewObjectMapStore(),
			serviceToSharedVipKeyStore:  NewObjectMapStore(),
		}
	})
	return lbinstance
}

type lbLister struct {
	lbStore *ObjectMapStore

	// annotationKey -> [svc1, svc2, svc3]
	sharedVipKeyToServicesStore *ObjectMapStore

	// svc1 -> annotationKey
	serviceToSharedVipKeyStore *ObjectMapStore
}

func (a *lbLister) Save(svcName string, lb interface{}) {
	utils.AviLog.Debugf("Saving lb svc :%s", svcName)
	a.lbStore.AddOrUpdate(svcName, lb)
}

func (a *lbLister) Get(svcName string) (bool, interface{}) {
	ok, obj := a.lbStore.Get(svcName)
	return ok, obj
}

func (a *lbLister) GetAll() interface{} {
	obj := a.lbStore.GetAllObjectNames()
	return obj
}

func (a *lbLister) Delete(svcName string) {
	a.lbStore.Delete(svcName)
}

func (a *lbLister) UpdateSharedVipKeyServiceMappings(key, svc string) {
	a.serviceToSharedVipKeyStore.AddOrUpdate(svc, key)
	found, services := a.GetSharedVipKeyToServices(key)
	if found {
		if utils.HasElem(services, svc) {
			return
		}
		services = append(services, svc)
		a.sharedVipKeyToServicesStore.AddOrUpdate(key, services)
		return
	}
	a.sharedVipKeyToServicesStore.AddOrUpdate(key, []string{svc})
}

func (a *lbLister) RemoveSharedVipKeyServiceMappings(svc string) bool {
	if found, key := a.GetServiceToSharedVipKey(svc); found {
		if foundServices, services := a.GetSharedVipKeyToServices(key); foundServices {
			services = utils.Remove(services, svc)
			if len(services) == 0 {
				a.sharedVipKeyToServicesStore.Delete(key)
			} else {
				a.sharedVipKeyToServicesStore.AddOrUpdate(key, services)
			}
		}
	}
	a.serviceToSharedVipKeyStore.Delete(svc)
	return true
}

func (a *lbLister) GetSharedVipKeyToServices(key string) (bool, []string) {
	found, serviceList := a.sharedVipKeyToServicesStore.Get(key)
	if !found {
		return false, make([]string, 0)
	}
	return true, serviceList.([]string)
}

func (a *lbLister) GetServiceToSharedVipKey(svc string) (bool, string) {
	found, key := a.serviceToSharedVipKeyStore.Get(svc)
	if !found {
		return false, ""
	}
	return true, key.(string)
}

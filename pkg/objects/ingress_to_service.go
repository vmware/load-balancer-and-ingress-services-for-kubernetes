/*
* [2013] - [2019] Avi Networks Incorporated
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

	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

//This package gives relationship APIs to manage a kubernetes service object.

var svclisterinstance *SvcLister
var svconce sync.Once

func SharedSvcLister() *SvcLister {
	svconce.Do(func() {
		svcIngStore := NewObjectStore()
		svclisterinstance = &SvcLister{}
		svclisterinstance.svcIngStore = svcIngStore
		ingSvcStore := NewObjectStore()
		svclisterinstance.ingSvcStore = ingSvcStore
	})
	return svclisterinstance
}

type SvcLister struct {
	svcIngStore *ObjectStore
	ingSvcStore *ObjectStore
}

type SvcNSCache struct {
	namespace     string
	svcIngobjects *ObjectMapStore
	IngressLock   sync.RWMutex
	IngNSCache
}

type IngNSCache struct {
	namespace     string
	ingSvcobjects *ObjectMapStore
}

func (v *SvcLister) IngressMappings(ns string) *SvcNSCache {
	namespacedsvcIngObjs := v.svcIngStore.GetNSStore(ns)
	namespacedIngSvcObjs := v.ingSvcStore.GetNSStore(ns)
	return &SvcNSCache{namespace: ns, svcIngobjects: namespacedsvcIngObjs, IngNSCache: IngNSCache{namespace: ns, ingSvcobjects: namespacedIngSvcObjs}}
}

//=====All service to ingress mapping methods are here.

func (v *SvcNSCache) GetSvcToIng(svcName string) (bool, []string) {
	// Need checks if it's found or not?
	found, ingNames := v.svcIngobjects.Get(svcName)
	if !found {
		return false, make([]string, 0)
	}
	return true, ingNames.([]string)
}

func (v *SvcNSCache) DeleteSvcToIngMapping(svcName string) bool {
	// Need checks if it's found or not?
	success := v.svcIngobjects.Delete(svcName)
	utils.AviLog.Info.Printf("Deleted the service mappings for svc: %s", svcName)
	return success
}

func (v *SvcNSCache) UpdateSvcToIngMapping(svcName string, ingressList []string) {
	utils.AviLog.Info.Printf("Updated the service mappings with svc: %s, ingresses: %s", svcName, ingressList)
	v.svcIngobjects.AddOrUpdate(svcName, ingressList)
}

//=====All ingress to service mapping methods are here.

func (v *IngNSCache) GetIngToSvc(ingName string) (bool, []string) {
	// Need checks if it's found or not?
	found, svcNames := v.ingSvcobjects.Get(ingName)
	if !found {
		return false, make([]string, 0)
	}
	return true, svcNames.([]string)
}

func (v *IngNSCache) DeleteIngToSvcMapping(ingName string) bool {
	// Need checks if it's found or not?
	success := v.ingSvcobjects.Delete(ingName)
	return success
}

func (v *IngNSCache) UpdateIngToSvcMapping(ingName string, svcList []string) {
	utils.AviLog.Info.Printf("Updated the ingress mappings with ingress: %s, svcs: %s", ingName, svcList)
	v.ingSvcobjects.AddOrUpdate(ingName, svcList)
}

func (v *IngNSCache) UpdatedIngressMappings(ingName string, svcList []string) {
	v.UpdateIngToSvcMapping(ingName, svcList)
}

//===All cross mapping update methods are here.

func (v *SvcNSCache) UpdateIngressMappings(ingName string, svcName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	_, ingresses := v.GetSvcToIng(svcName)
	ingresses = append(ingresses, ingName)
	v.UpdateSvcToIngMapping(svcName, ingresses)
	_, svcs := v.GetIngToSvc(ingName)
	svcs = append(svcs, svcName)
	v.UpdateIngToSvcMapping(ingName, svcs)
}

func (v *SvcNSCache) RemoveIngressMappings(ingName string) {
	v.IngressLock.Lock()
	defer v.IngressLock.Unlock()
	// Get all the services for the ingress
	ok, svcs := v.GetIngToSvc(ingName)
	// Iterate and remove this ingress from the service mappings
	if ok {
		for _, svc := range svcs {
			found, ingresses := v.GetSvcToIng(svc)
			if found {
				ingresses = Remove(ingresses, ingName)
				// Update the service mapping
				v.UpdateSvcToIngMapping(svc, ingresses)
			}
		}
	}
	// Remove the ingress from the ingress --> service map
	v.DeleteIngToSvcMapping(ingName)
}

// Candidate for utils package
func Remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

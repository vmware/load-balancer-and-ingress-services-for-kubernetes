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

import "sync"

//This package gives relationship APIs to manage a kubernetes service object.

var svclisterinstance *SvcLister
var svconce sync.Once

func SharedSvcLister() *SvcLister {
	svconce.Do(func() {
		svcIngStore := NewObjectStore()
		svclisterinstance = &SvcLister{}
		svclisterinstance.svcIngStore = svcIngStore
	})
	return svclisterinstance
}

type SvcLister struct {
	svcIngStore *ObjectStore
}

type SvcNSCache struct {
	namespace     string
	svcIngobjects *ObjectMapStore
}

func (v *SvcLister) Service(ns string) *SvcNSCache {
	namespacedsvcIngObjs := v.svcIngStore.GetNSStore(ns)
	return &SvcNSCache{namespace: ns, svcIngobjects: namespacedsvcIngObjs}
}

func (v *SvcNSCache) GetSvcToIng(svcName string) (bool, []string) {
	// Need checks if it's found or not?
	found, ingNames := v.svcIngobjects.Get(svcName)
	if !found {
		return false, nil
	}
	return true, ingNames.([]string)
}

func (v *SvcNSCache) DeleteSvcToIngMapping(svcName string) bool {
	// Need checks if it's found or not?
	success := v.svcIngobjects.Delete(svcName)
	return success
}

func (v *SvcNSCache) UpdateSvcToIngMapping(svcName string, ingressList []string) {
	v.svcIngobjects.AddOrUpdate(svcName, ingressList)
}

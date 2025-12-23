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

// APIs to access AVI graph from and to memory. The VS should have a uuid and the corresponding model.

package objects

import (
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var aviGraphinstance *AviGraphLister
var avionce sync.Once

func SharedAviGraphLister() *AviGraphLister {
	avionce.Do(func() {
		AviGraphStore := NewObjectMapStore()
		aviGraphinstance = &AviGraphLister{}
		aviGraphinstance.AviGraphStore = AviGraphStore
	})
	return aviGraphinstance
}

type AviGraphLister struct {
	AviGraphStore *ObjectMapStore
}

func (a *AviGraphLister) Save(vsName string, aviGraph interface{}) {
	utils.AviLog.Infof("Saving Model: %s", vsName)
	a.AviGraphStore.AddOrUpdate(vsName, aviGraph)
}

func (a *AviGraphLister) Get(vsName string) (bool, interface{}) {
	ok, obj := a.AviGraphStore.Get(vsName)
	return ok, obj
}

func (a *AviGraphLister) GetAll() interface{} {
	obj := a.AviGraphStore.GetAllObjectNames()
	return obj
}

func (a *AviGraphLister) Delete(vsName string) {
	a.AviGraphStore.Delete(vsName)
}

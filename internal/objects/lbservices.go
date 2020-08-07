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

	"ako/pkg/utils"
)

var lbinstance *lbLister
var lbonce sync.Once

func SharedlbLister() *lbLister {
	lbonce.Do(func() {
		lbStore := NewObjectMapStore()
		lbinstance = &lbLister{}
		lbinstance.lbStore = lbStore
	})
	return lbinstance
}

type lbLister struct {
	lbStore *ObjectMapStore
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

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

// APIs to access AVI graph from and to memory. The VS should have a uuid and the corresponding model.

package objects

import (
	"sync"
)

var resouceVerInstance *ResourceVersionLister
var resourceVerOnce sync.Once

func SharedResourceVerInstanceLister() *ResourceVersionLister {
	resourceVerOnce.Do(func() {
		ResourceVerStore := NewObjectMapStore()
		resouceVerInstance = &ResourceVersionLister{}
		resouceVerInstance.ResourceVerStore = ResourceVerStore
	})
	return resouceVerInstance
}

type ResourceVersionLister struct {
	ResourceVerStore *ObjectMapStore
}

func (a *ResourceVersionLister) Save(vsName string, resVer interface{}) {
	a.ResourceVerStore.AddOrUpdate(vsName, resVer)
}

func (a *ResourceVersionLister) Get(resName string) (bool, interface{}) {
	ok, obj := a.ResourceVerStore.Get(resName)
	return ok, obj
}

func (a *ResourceVersionLister) GetAll() interface{} {
	obj := a.ResourceVerStore.GetAllObjectNames()
	return obj
}

func (a *ResourceVersionLister) Delete(resName string) {
	a.ResourceVerStore.Delete(resName)

}

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
)

var nsListerInstace *NamespaceLister
var nsListerOnce sync.Once

func SharedNamespaceTenantLister() *NamespaceLister {
	nsListerOnce.Do(func() {
		nsListerInstace = &NamespaceLister{
			NamespacedResourceToTenantStore: NewObjectMapStore(),
		}
	})
	return nsListerInstace
}

type NamespaceLister struct {
	NamespaceLock sync.RWMutex
	//namespace --> tenant
	NamespacedResourceToTenantStore *ObjectMapStore
}

func (n *NamespaceLister) UpdateNamespacedResourceToTenantStore(key, tenant string) {
	n.NamespaceLock.Lock()
	defer n.NamespaceLock.Unlock()
	n.NamespacedResourceToTenantStore.AddOrUpdate(key, tenant)
}

func (n *NamespaceLister) GetTenantInNamespace(key string) string {
	found, tenant := n.NamespacedResourceToTenantStore.Get(key)
	if !found {
		return ""
	}
	return tenant.(string)
}

func (n *NamespaceLister) RemoveNamespaceToTenantCache(key string) bool {
	n.NamespaceLock.Lock()
	defer n.NamespaceLock.Unlock()
	return n.NamespacedResourceToTenantStore.Delete(key)
}

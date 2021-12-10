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

// Construct in memory database that populates updates from both kubernetes and MCP
// The format is: namespace:[object_name: obj]

package objects

import (
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type ObjectStore struct {
	NSObjectMap map[string]*ObjectMapStore
	NSLock      sync.RWMutex
}

func NewObjectStore() *ObjectStore {
	objectStore := &ObjectStore{}
	objectStore.NSObjectMap = make(map[string]*ObjectMapStore)
	return objectStore
}

func (store *ObjectStore) GetNSStore(nsName string) *ObjectMapStore {
	store.NSLock.Lock()
	defer store.NSLock.Unlock()
	val, ok := store.NSObjectMap[nsName]
	if ok {
		return val
	} else {
		// This namespace is not initialized, let's initialze it
		nsObjStore := NewObjectMapStore()
		// Update the store.
		store.NSObjectMap[nsName] = nsObjStore
		return nsObjStore
	}
}

func (store *ObjectStore) DeleteNSStore(nsName string) bool {
	// Deletes the key for a namespace. Wipes off the entire NS. So use with care.
	store.NSLock.Lock()
	defer store.NSLock.Unlock()
	_, ok := store.NSObjectMap[nsName]
	if ok {
		delete(store.NSObjectMap, nsName)
		return true
	}
	utils.AviLog.Warnf("Namespace: %s not found, nothing to delete returning false", nsName)
	return false

}

func (store *ObjectStore) GetAllNamespaces() []string {
	// Take a read lock on the store and write lock on NS object
	store.NSLock.RLock()
	defer store.NSLock.RUnlock()
	var allNamespaces []string
	for ns := range store.NSObjectMap {
		allNamespaces = append(allNamespaces, ns)
	}
	return allNamespaces

}

type ObjectMapStore struct {
	ObjectMap map[string]interface{}
	ObjLock   sync.RWMutex
}

func NewObjectMapStore() *ObjectMapStore {
	nsObjStore := &ObjectMapStore{}
	nsObjStore.ObjectMap = make(map[string]interface{})
	return nsObjStore
}

func (o *ObjectMapStore) AddOrUpdate(objName string, obj interface{}) {
	o.ObjLock.Lock()
	defer o.ObjLock.Unlock()
	o.ObjectMap[objName] = obj
}

func (o *ObjectMapStore) Delete(objName string) bool {
	o.ObjLock.Lock()
	defer o.ObjLock.Unlock()
	_, ok := o.ObjectMap[objName]
	if ok {
		delete(o.ObjectMap, objName)
		return true
	}
	return false

}

func (o *ObjectMapStore) Get(objName string) (bool, interface{}) {
	o.ObjLock.RLock()
	defer o.ObjLock.RUnlock()
	val, ok := o.ObjectMap[objName]
	if ok {
		return true, val
	}
	return false, nil

}

func (o *ObjectMapStore) GetAllObjectNames() map[string]interface{} {
	o.ObjLock.RLock()
	defer o.ObjLock.RUnlock()
	// TODO (sudswas): Pass a copy instead of the reference
	return o.ObjectMap

}

func (o *ObjectMapStore) GetAllKeys() []string {
	o.ObjLock.RLock()
	defer o.ObjLock.RUnlock()
	allKeys := []string{}
	for k := range o.ObjectMap {
		allKeys = append(allKeys, k)
	}
	return allKeys
}

func (o *ObjectMapStore) CopyAllObjects() map[string]interface{} {
	o.ObjLock.RLock()
	defer o.ObjLock.RUnlock()
	CopiedObjMap := make(map[string]interface{})
	for k, v := range o.ObjectMap {
		CopiedObjMap[k] = v
	}
	return CopiedObjMap
}

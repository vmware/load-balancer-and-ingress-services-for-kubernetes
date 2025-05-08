/*
 * Copyright 2020-2021 VMware, Inc.
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

var nplInstance *NPLLister
var nplOnce sync.Once

func SharedNPLLister() *NPLLister {
	nplOnce.Do(func() {
		store := NewObjectMapStore()
		nplInstance = &NPLLister{}
		nplInstance.store = store
	})
	return nplInstance
}

// NPLLister stores a list of NPL annotations for a pod
type NPLLister struct {
	store *ObjectMapStore
}

func (a *NPLLister) Save(key string, val interface{}) {
	a.store.AddOrUpdate(key, val)
}

func (a *NPLLister) Get(key string) (bool, interface{}) {
	ok, obj := a.store.Get(key)
	return ok, obj
}

func (a *NPLLister) GetAll() interface{} {
	obj := a.store.GetAllObjectNames()
	return obj
}

func (a *NPLLister) Delete(key string) {
	a.store.Delete(key)

}

var podSvcInstance *PodSvcLister
var podSvcOnce sync.Once

func SharedPodToSvcLister() *PodSvcLister {
	podSvcOnce.Do(func() {
		store := NewObjectMapStore()
		podSvcInstance = &PodSvcLister{}
		podSvcInstance.store = store
	})
	return podSvcInstance
}

// PodSvcLister stores list of services for a pod.
// For all these services, the Pod acts a backend through matching selector
type PodSvcLister struct {
	store *ObjectMapStore
}

func (a *PodSvcLister) Save(key string, val interface{}) {
	a.store.AddOrUpdate(key, val)
}

func (a *PodSvcLister) Get(key string) (bool, interface{}) {
	ok, obj := a.store.Get(key)
	return ok, obj
}

func (a *PodSvcLister) GetAll() interface{} {
	obj := a.store.GetAllObjectNames()
	return obj
}

func (a *PodSvcLister) Delete(key string) {
	a.store.Delete(key)
}

var svcPodInstance *SvcPodLister
var svcPodOnce sync.Once

func SharedSvcToPodLister() *SvcPodLister {
	svcPodOnce.Do(func() {
		store := NewObjectMapStore()
		svcPodInstance = &SvcPodLister{}
		svcPodInstance.store = store
	})
	return svcPodInstance
}

// SvcPodLister stores list of pods for a service with matching label selector
type SvcPodLister struct {
	store *ObjectMapStore
}

func (a *SvcPodLister) Save(key string, val interface{}) {
	a.store.AddOrUpdate(key, val)
}

func (a *SvcPodLister) Get(key string) (bool, interface{}) {
	ok, obj := a.store.Get(key)
	return ok, obj
}

func (a *SvcPodLister) GetAll() interface{} {
	obj := a.store.GetAllObjectNames()
	return obj
}

func (a *SvcPodLister) Delete(key string) {
	a.store.Delete(key)
}

var podLBSvcInstance *PodLBSvcLister
var podLBSvcOnce sync.Once

func SharedPodToLBSvcLister() *PodLBSvcLister {
	podLBSvcOnce.Do(func() {
		store := NewObjectMapStore()
		podLBSvcInstance = &PodLBSvcLister{}
		podLBSvcInstance.store = store
	})
	return podLBSvcInstance
}

// PodLBSvcLister stores list of services of type LB for a pod.
type PodLBSvcLister struct {
	store *ObjectMapStore
}

func (a *PodLBSvcLister) Save(key string, val interface{}) {
	a.store.AddOrUpdate(key, val)
}

func (a *PodLBSvcLister) Get(key string) (bool, interface{}) {
	ok, obj := a.store.Get(key)
	return ok, obj
}

func (a *PodLBSvcLister) GetAll() interface{} {
	obj := a.store.GetAllObjectNames()
	return obj
}

func (a *PodLBSvcLister) Delete(key string) {
	a.store.Delete(key)
}

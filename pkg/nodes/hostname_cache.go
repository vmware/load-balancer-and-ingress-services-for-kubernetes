/*
* [2013] - [2020] Avi Networks Incorporated
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

package nodes

import (
	"ako/pkg/objects"
	"sync"

	"github.com/avinetworks/container-lib/utils"
)

var hostNameLister *HostNameLister
var hsonce sync.Once

func SharedHostNameLister() *HostNameLister {
	hsonce.Do(func() {
		hostNameLister = &HostNameLister{
			secureHostNameStore: objects.NewObjectMapStore(),
			HostNamePathStore: HostNamePathStore{
				hostNamePathStore: objects.NewObjectMapStore(),
			},
		}
	})
	return hostNameLister
}

type HostNameLister struct {
	secureHostNameStore *objects.ObjectMapStore
	HostNamePathStore
}

func (a *HostNameLister) Save(hostname string, hsGraph SecureHostNameMapProp) {
	utils.AviLog.Infof("Saving hostname map :%s", hostname)
	a.secureHostNameStore.AddOrUpdate(hostname, hsGraph)
}

func (a *HostNameLister) Get(hostname string) (bool, SecureHostNameMapProp) {
	ok, obj := a.secureHostNameStore.Get(hostname)
	if !ok {
		return ok, SecureHostNameMapProp{}
	}
	return ok, obj.(SecureHostNameMapProp)
}

func (a *HostNameLister) Delete(hostname string) {
	a.secureHostNameStore.Delete(hostname)
}

// thread safe for namespace based sharding in case of same hostname in different namespaces
// cache sample: foo.com/path1 -> [ns1/ingress1]
type HostNamePathStore struct {
	sync.RWMutex
	hostNamePathStore *objects.ObjectMapStore
}

func (h *HostNamePathStore) GetHostPathStore(hostpath string) (bool, []string) {
	ok, obj := h.hostNamePathStore.Get(hostpath)
	if !ok {
		return false, []string{}
	}
	return true, obj.([]string)
}

func (h *HostNamePathStore) SaveHostPathStore(hostpath string, data string) {
	h.Lock()
	defer h.Unlock()
	found, obj := h.GetHostPathStore(hostpath)
	if found && !utils.HasElem(obj, data) {
		obj = append(obj, data)
	} else {
		obj = []string{data}
	}
	h.hostNamePathStore.AddOrUpdate(hostpath, obj)
}

func (h *HostNamePathStore) RemoveHostPathStore(hostpath string, data string) {
	h.Lock()
	defer h.Unlock()
	found, obj := h.GetHostPathStore(hostpath)
	if found && utils.HasElem(obj, data) {
		obj = utils.Remove(obj, data)
		h.hostNamePathStore.AddOrUpdate(hostpath, obj)
	}

	if len(obj) == 0 {
		h.DeleteHostPathStore(hostpath)
	}
}

func (h *HostNamePathStore) DeleteHostPathStore(hostpath string) {
	h.hostNamePathStore.Delete(hostpath)
}

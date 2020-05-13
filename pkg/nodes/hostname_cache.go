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
		HostNameStore := objects.NewObjectMapStore()
		hostNameLister = &HostNameLister{}
		hostNameLister.HostNameStore = HostNameStore
	})
	return hostNameLister
}

type HostNameLister struct {
	HostNameStore *objects.ObjectMapStore
}

func (a *HostNameLister) Save(hostname string, hsGraph SecureHostNameMapProp) {
	utils.AviLog.Infof("Saving hostname map :%s", hostname)
	a.HostNameStore.AddOrUpdate(hostname, hsGraph)
}

func (a *HostNameLister) Get(hostname string) (bool, SecureHostNameMapProp) {
	ok, obj := a.HostNameStore.Get(hostname)
	if !ok {
		return ok, SecureHostNameMapProp{}
	}
	return ok, obj.(SecureHostNameMapProp)
}

func (a *HostNameLister) Delete(hostname string) {
	a.HostNameStore.Delete(hostname)

}

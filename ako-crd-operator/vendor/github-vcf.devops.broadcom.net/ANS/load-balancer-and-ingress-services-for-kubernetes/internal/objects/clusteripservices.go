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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var clusterIpinstance *clusterIpLister
var clusterIponce sync.Once

func SharedClusterIpLister() *clusterIpLister {
	clusterIponce.Do(func() {
		clusterIpStore := NewObjectMapStore()
		clusterIpinstance = &clusterIpLister{}
		clusterIpinstance.clusterIpStore = clusterIpStore
	})
	return clusterIpinstance
}

type clusterIpLister struct {
	clusterIpStore *ObjectMapStore
}

func (a *clusterIpLister) Save(svcName string, lb interface{}) {
	utils.AviLog.Debugf("Saving clusterIp svc :%s", svcName)
	a.clusterIpStore.AddOrUpdate(svcName, lb)
}

func (a *clusterIpLister) Get(svcName string) (bool, interface{}) {
	ok, obj := a.clusterIpStore.Get(svcName)
	return ok, obj
}

func (a *clusterIpLister) GetAll() interface{} {
	obj := a.clusterIpStore.GetAllObjectNames()
	return obj
}

func (a *clusterIpLister) Delete(svcName string) {
	a.clusterIpStore.Delete(svcName)

}

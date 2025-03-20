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

package nodes

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (o *AviObjectGraph) BuildAppProfileGraph(ns, name, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	aviAppProfNode := &AviAppProfileNode{
		Name:                name,
		Tenant:              ns,
		Type:                lib.AllowedL4ApplicationProfile,
		EnableProxyProtocol: true,
	}
	o.AddModelNode(aviAppProfNode)
	aviAppProfNode.CalculateCheckSum()
	utils.AviLog.Infof("key: %s, Added app profile node %s", key, name)
	utils.AviLog.Debugf("key: %s, app profile node: [%v]", key, utils.Stringify(aviAppProfNode))
}

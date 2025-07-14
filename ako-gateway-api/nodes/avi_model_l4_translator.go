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

package nodes

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
)

func (o *AviObjectGraph) ProcessL4Routes(key string, routeModel RouteModel, parentNsName string) {
	for _, rule := range routeModel.ParseRouteConfig(key).Rules {
		parentNode := o.GetAviEvhVS()
		// create L4 policyset per rule
		o.BuildL4PolicySet(key, parentNode[0], routeModel, rule)
		// Logic to override the app profile per port also must be taken care here.
	}
}

func (o *AviObjectGraph) BuildL4PolicySet(key string, vsNode *nodes.AviEvhVsNode, routeModel RouteModel, rule *Rule) {

	// TODO: add the l4policset code here
}

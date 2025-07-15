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

package lib

const (
	Prefix            = "ako-gw-"
	GatewayController = "ako.vmware.com/avi-lb"
	CoreGroup         = "v1"
	GatewayGroup      = "gateway.networking.k8s.io"
	HealthMonitorKind = "HealthMonitor"
)

const (
	ZeroAttachedRoutes = 0
)

const (
	GatewayClassGatewayControllerIndex = "GatewayClassGatewayController"
	REGULAREXPRESSION                  = "RegularExpression"
	EXACT                              = "Exact"
	PATHPREFIX                         = "PathPrefix"
	LBVipTypeAnnotation                = "networking.vmware.com/lb-vip-type"
	VCFGatewayClassName                = "avi-lb"
)

const (
	AllowedRoutesNamespaceFromAll  = "All"
	AllowedRoutesNamespaceFromSame = "Same"
)

var SupportedLBVipTypes = map[string]string{
	"public":  "PUBLIC",
	"private": "PRIVATE",
}

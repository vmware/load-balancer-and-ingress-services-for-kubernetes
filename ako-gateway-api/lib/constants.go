/*
 * Copyright 2023-2024 VMware, Inc.
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
	Prefix                         = "ako-gw-"
	GatewayController              = "ako.vmware.com/avi-lb"
	CoreGroup                      = "v1"
	GatewayGroup                   = "gateway.networking.k8s.io"
	BackendRefFilterDatascriptName = "BackendRefFilterDatascript"
	AddHeaderStringGroup           = "AddHeaderStringGroup"
	UpdateHeaderStringGroup        = "UpdateHeaderStringGroup"
	DeleteHeaderStringGroup        = "DeleteHeaderStringGroup"
)

const (
	ZeroAttachedRoutes = 0
)

const (
	GatewayClassGatewayControllerIndex = "GatewayClassGatewayController"
)

const (
	AllowedRoutesNamespaceFromAll  = "All"
	AllowedRoutesNamespaceFromSame = "Same"
)

// BackendRefFIlter data script is used to support Filters within backendRef in a HTTPRoute in Gateways
const (
	BackendRefFilterDatascript = `pool_name = avi.pool.name()
  	add_header_strgrp = "NAMEPREFIX-AddHeaderStringGroup"
	val, match_found = avi.stringgroup.contains(add_header_strgrp, pool_name)
	if match_found then
   		for header_string in string.gmatch(val, '([^,]+)') do
      		local colon_index=string.find(header_string,":")
      		local headerkey = string.sub(header_string,1,colon_index-1)
      		local headerval = string.sub(header_string,colon_index+1)
      		avi.http.add_header(headerkey, headerval)
   		end
	end
	update_header_strgrp = "NAMEPREFIX-UpdateHeaderStringGroup"
	val, match_found = avi.stringgroup.contains(update_header_strgrp, pool_name)
	if match_found then
   		for header_string in string.gmatch(val, '([^,]+)') do
      		local colon_index=string.find(header_string,":")
      		local headerkey = string.sub(header_string,1,colon_index-1)
      		local headerval = string.sub(header_string,colon_index+1)
      		avi.http.replace_header(headerkey, headerval)
   		end
	end
	delete_header_strgrp = "NAMEPREFIX-DeleteHeaderStringGroup"
	val, match_found = avi.stringgroup.contains(delete_header_strgrp, pool_name)
	if match_found then
   		for header_key in string.gmatch(val, '([^,]+)') do
      		avi.http.remove_header(header_key)
   		end
	end`
)

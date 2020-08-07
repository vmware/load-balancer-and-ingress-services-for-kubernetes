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

package lib

import (
	"ako/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
)

func StaticRoutesIntfToObj(staticRoutesIntf []interface{}) []*avimodels.StaticRoute {
	var staticRoutes []*avimodels.StaticRoute

	for _, staticRouteIf := range staticRoutesIntf {
		staticRouteMap, ok := staticRouteIf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warnf("object type %T did not match for StaticRoute\n", staticRouteIf)
			continue
		}
		staticRoute := avimodels.StaticRoute{}
		for key, val := range staticRouteMap {
			switch key {
			case "disable_gateway_monitor":
				if disableGatewayMonitor, ok := val.(bool); ok {
					staticRoute.DisableGatewayMonitor = &disableGatewayMonitor
				} else {
					utils.AviLog.Warnf("wrong object type %T for disableGatewayMonitor in staticRoute\n", val)
				}
			case "if_name":
				if ifName, ok := val.(string); ok {
					staticRoute.IfName = &ifName
				} else {
					utils.AviLog.Warnf("wrong object type %T for ifName in staticRoute\n", val)
				}
			case "next_hop":
				if nextHop, ok := val.(map[string]interface{}); ok {
					staticRoute.NextHop = IPAddrIntfToObj(nextHop)
				} else {
					utils.AviLog.Warnf("wrong object type %T for nextHop in staticRoute\n", val)
				}
			case "prefix":
				if prefix, ok := val.(map[string]interface{}); ok {
					staticRoute.Prefix = IAddrPrefixIntfToObj(prefix)
				} else {
					utils.AviLog.Warnf("wrong object type %T for prefix in staticRoute\n", val)
				}
			case "route_id":
				if routeId, ok := val.(string); ok {
					staticRoute.RouteID = &routeId
				} else {
					utils.AviLog.Warnf("wrong object type %T for routeId in staticRoute\n", val)
				}
			default:
				utils.AviLog.Warnf("Unknown key %s in staticRoute\n", key)
			}
		}
		staticRoutes = append(staticRoutes, &staticRoute)
	}
	return staticRoutes
}

func IPAddrIntfToObj(ipAddrIntf map[string]interface{}) *avimodels.IPAddr {
	ipAddr := avimodels.IPAddr{}
	for key, val := range ipAddrIntf {
		switch key {
		case "addr":
			if addr, ok := val.(string); ok {
				ipAddr.Addr = &addr
			} else {
				utils.AviLog.Warnf("wrong object type %T for addr in IPAddr\n", val)
			}
		case "type":
			if addrType, ok := val.(string); ok {
				ipAddr.Type = &addrType
			} else {
				utils.AviLog.Warnf("wrong object type %T for type in IPAddr\n", val)
			}
		default:
			utils.AviLog.Warnf("Unknown key %s in IPAddr\n", key)
		}
	}
	return &ipAddr
}

func IAddrPrefixIntfToObj(ipAddrPrefixIntf map[string]interface{}) *avimodels.IPAddrPrefix {
	ipAddrPrefix := avimodels.IPAddrPrefix{}
	for key, val := range ipAddrPrefixIntf {
		switch key {
		case "ip_addr":
			if ipaddr, ok := val.(map[string]interface{}); ok {
				ipAddrPrefix.IPAddr = IPAddrIntfToObj(ipaddr)
			} else {
				utils.AviLog.Warnf("wrong object type %T for IPAddr in IPAddrPrefix\n", val)
			}
		case "mask":
			if mask, ok := val.(float64); ok {
				mask32 := int32(mask)
				ipAddrPrefix.Mask = &mask32
			} else {
				utils.AviLog.Warnf("wrong object type %T for Mask in IPAddrPrefix\n", val)
			}
		default:
			utils.AviLog.Warnf("Unknown key %s in IPAddrPefix\n", key)
		}
	}
	return &ipAddrPrefix
}

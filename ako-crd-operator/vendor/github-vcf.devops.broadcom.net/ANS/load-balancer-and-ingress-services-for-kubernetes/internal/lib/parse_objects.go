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

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"
)

func StaticRoutesIntfToObj(staticRoutesIntf []interface{}) []*avimodels.StaticRoute {
	var staticRoutes []*avimodels.StaticRoute

	for _, staticRouteIf := range staticRoutesIntf {
		staticRouteMap, ok := staticRouteIf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warnf("object type %T did not match for StaticRoute", staticRouteIf)
			continue
		}
		staticRoute := avimodels.StaticRoute{}
		for key, val := range staticRouteMap {
			switch key {
			case "disable_gateway_monitor":
				if disableGatewayMonitor, ok := val.(bool); ok {
					staticRoute.DisableGatewayMonitor = &disableGatewayMonitor
				} else {
					utils.AviLog.Warnf("wrong object type %T for disableGatewayMonitor in staticRoute", val)
				}
			case "if_name":
				if ifName, ok := val.(string); ok {
					staticRoute.IfName = &ifName
				} else {
					utils.AviLog.Warnf("wrong object type %T for ifName in staticRoute", val)
				}
			case "next_hop":
				if nextHop, ok := val.(map[string]interface{}); ok {
					staticRoute.NextHop = IPAddrIntfToObj(nextHop)
				} else {
					utils.AviLog.Warnf("wrong object type %T for nextHop in staticRoute", val)
				}
			case "prefix":
				if prefix, ok := val.(map[string]interface{}); ok {
					staticRoute.Prefix = IAddrPrefixIntfToObj(prefix)
				} else {
					utils.AviLog.Warnf("wrong object type %T for prefix in staticRoute", val)
				}
			case "route_id":
				if routeId, ok := val.(string); ok {
					staticRoute.RouteID = &routeId
				} else {
					utils.AviLog.Warnf("wrong object type %T for routeId in staticRoute", val)
				}
			case "labels":
				if labels, ok := val.([]interface{}); ok {
					staticRoute.Labels = LabelsIntfToObj(labels)
				} else {
					utils.AviLog.Warnf("wrong object type %T for labels in staticRoute", val)
				}
			default:
				utils.AviLog.Warnf("Unknown key %s in staticRoute", key)
			}
		}
		staticRoutes = append(staticRoutes, &staticRoute)
	}
	return staticRoutes
}

func LabelsIntfToObj(labelsIntf []interface{}) []*avimodels.KeyValue {
	var labels []*avimodels.KeyValue

	for _, labelIntf := range labelsIntf {
		labelMap, ok := labelIntf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warnf("object type %T did not match for label", labelIntf)
			continue
		}
		kv := keyValueIntfToObj(labelMap)
		labels = append(labels, kv)

	}
	return labels
}

func keyValueIntfToObj(keyValueIntf map[string]interface{}) *avimodels.KeyValue {
	keyValue := avimodels.KeyValue{}
	for key, val := range keyValueIntf {
		switch key {
		case "key":
			if k, ok := val.(string); ok {
				keyValue.Key = &k
			} else {
				utils.AviLog.Warnf("wrong object type %T for addr in KeyValue", val)
			}
		case "value":
			if v, ok := val.(string); ok {
				keyValue.Value = &v
			} else {
				utils.AviLog.Warnf("wrong object type %T for type in KeyValue", val)
			}
		default:
			utils.AviLog.Warnf("Unknown key %s in KeyValue", key)
		}
	}
	return &keyValue
}

func IPAddrIntfToObj(ipAddrIntf map[string]interface{}) *avimodels.IPAddr {
	ipAddr := avimodels.IPAddr{}
	for key, val := range ipAddrIntf {
		switch key {
		case "addr":
			if addr, ok := val.(string); ok {
				ipAddr.Addr = &addr
			} else {
				utils.AviLog.Warnf("wrong object type %T for addr in IPAddr", val)
			}
		case "type":
			if addrType, ok := val.(string); ok {
				ipAddr.Type = &addrType
			} else {
				utils.AviLog.Warnf("wrong object type %T for type in IPAddr", val)
			}
		default:
			utils.AviLog.Warnf("Unknown key %s in IPAddr", key)
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
				utils.AviLog.Warnf("wrong object type %T for IPAddr in IPAddrPrefix", val)
			}
		case "mask":
			if mask, ok := val.(float64); ok {
				mask32 := int32(mask)
				ipAddrPrefix.Mask = &mask32
			} else {
				utils.AviLog.Warnf("wrong object type %T for Mask in IPAddrPrefix", val)
			}
		default:
			utils.AviLog.Warnf("Unknown key %s in IPAddrPefix", key)
		}
	}
	return &ipAddrPrefix
}

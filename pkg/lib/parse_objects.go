/*
* [2013] - [2019] Avi Networks Incorporated
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
	avimodels "github.com/avinetworks/sdk/go/models"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

func StaticRoutesIntfToObj(staticRoutesIntf []interface{}) []*avimodels.StaticRoute {
	var staticRoutes []*avimodels.StaticRoute

	for _, staticRouteIf := range staticRoutesIntf {
		staticRouteMap, ok := staticRouteIf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("object type %T did not match for StaticRoute\n", staticRouteIf)
			continue
		}
		staticRoute := avimodels.StaticRoute{}
		for key, val := range staticRouteMap {
			switch key {
			case "disable_gateway_monitor":
				if disableGatewayMonitor, ok := val.(bool); ok {
					staticRoute.DisableGatewayMonitor = &disableGatewayMonitor
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for disableGatewayMonitor in staticRoute\n", val)
				}
			case "if_name":
				if ifName, ok := val.(string); ok {
					staticRoute.IfName = &ifName
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for ifName in staticRoute\n", val)
				}
			case "next_hop":
				if nextHop, ok := val.(map[string]interface{}); ok {
					staticRoute.NextHop = IPAddrIntfToObj(nextHop)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for nextHop in staticRoute\n", val)
				}
			case "prefix":
				if prefix, ok := val.(map[string]interface{}); ok {
					staticRoute.Prefix = IAddrPrefixIntfToObj(prefix)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for prefix in staticRoute\n", val)
				}
			case "route_id":
				if routeId, ok := val.(string); ok {
					staticRoute.RouteID = &routeId
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for routeId in staticRoute\n", val)
				}
			default:
				utils.AviLog.Warning.Printf("Unknown key %s in staticRoute\n", key)
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
				utils.AviLog.Warning.Printf("wrong object type %T for addr in IPAddr\n", val)
			}
		case "type":
			if addrType, ok := val.(string); ok {
				ipAddr.Type = &addrType
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for type in IPAddr\n", val)
			}
		default:
			utils.AviLog.Warning.Printf("Unknown key %s in IPAddr\n", key)
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
				utils.AviLog.Warning.Printf("wrong object type %T for IPAddr in IPAddrPrefix\n", val)
			}
		case "mask":
			if mask, ok := val.(float64); ok {
				mask32 := int32(mask)
				ipAddrPrefix.Mask = &mask32
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for Mask in IPAddrPrefix\n", val)
			}
		default:
			utils.AviLog.Warning.Printf("Unknown key %s in IPAddrPefix\n", key)
		}
	}
	return &ipAddrPrefix
}

func IpCommunitiesIntfToObj(ipCommunitiesIntf []interface{}) []*avimodels.IPCommunity {
	var ipCommunities []*avimodels.IPCommunity
	for _, ipcIntf := range ipCommunitiesIntf {
		ipcMap, ok := ipcIntf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("ipCommunities not of correct type\n")
			continue
		}
		ipCommunity := avimodels.IPCommunity{}
		for key, val := range ipcMap {
			switch key {
			case "community":
				if community, ok := val.([]string); ok {
					ipCommunity.Community = community
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for Community in IpCommunity\n", val)
				}
			case "ip_begin":
				if ipAddrIntf, ok := val.(map[string]interface{}); ok {
					ipCommunity.IPBegin = IPAddrIntfToObj(ipAddrIntf)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for IPBegin in IpCommunity\n", val)
				}
			case "ip_end":
				if ipAddrIntf, ok := val.(map[string]interface{}); ok {
					ipCommunity.IPBegin = IPAddrIntfToObj(ipAddrIntf)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for IPEnd in IpCommunity\n", val)
				}
			default:
				utils.AviLog.Warning.Printf("Unknown key %s in IpCommunity\n", key)
			}
		}
		ipCommunities = append(ipCommunities, &ipCommunity)
	}
	return ipCommunities
}

func bgpPeerIntfToObj(bgpPeerIntf []interface{}) []*avimodels.BgpPeer {
	var bgpPeers []*avimodels.BgpPeer
	for _, bgpIntf := range bgpPeerIntf {
		bgpMapIntf, ok := bgpIntf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("object type %T did not match for bgpPeer\n", bgpIntf)
			continue
		}
		bgpPeer := avimodels.BgpPeer{}
		for key, val := range bgpMapIntf {
			switch key {
			case "advertise_snat_ip":
				if adSnatIP, ok := val.(bool); ok {
					bgpPeer.AdvertiseSnatIP = &adSnatIP
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for AdvertiseSnatIp in BgpPeer\n", val)
				}
			case "advertise_vip":
				if advVip, ok := val.(bool); ok {
					bgpPeer.AdvertiseVip = &advVip
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for AdvertiseVip in BgpPeer\n", val)
				}
			case "advertise_interval":
				if adInterval, ok := val.(int32); ok {
					bgpPeer.AdvertisementInterval = &adInterval
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for AdvertiseInterval in BgpPeer\n", val)
				}
			case "bfd":
				if bfd, ok := val.(bool); ok {
					bgpPeer.Bfd = &bfd
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for Bfd in BgpPeer\n", val)
				}
			case "connection_timer":
				if connectionTimer, ok := val.(int32); ok {
					bgpPeer.AdvertisementInterval = &connectionTimer
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for ConnectionTimer in BgpPeer\n", val)
				}
			case "ebgp_multihop":
				if ebgpMultihop, ok := val.(int32); ok {
					bgpPeer.AdvertisementInterval = &ebgpMultihop
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for EbgpMultihop in BgpPeer\n", val)
				}
			case "hold_time":
				if holdTime, ok := val.(int32); ok {
					bgpPeer.LocalAs = &holdTime
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for HoldTime in BgpPeer\n", val)
				}
			case "keep_alive_interval":
				if keepAliveInterval, ok := val.(int32); ok {
					bgpPeer.LocalAs = &keepAliveInterval
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for KeppAliveInterval in BgpPeer\n", val)
				}
			case "local_as":
				if localAs, ok := val.(int32); ok {
					bgpPeer.LocalAs = &localAs
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for LocalAs in BgpPeer\n", val)
				}
			case "md5_secret":
				if md5Secret, ok := val.(string); ok {
					bgpPeer.Md5Secret = &md5Secret
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for Md5Secret in BgpPeer\n", val)
				}
			case "network_ref":
				if networkRef, ok := val.(string); ok {
					bgpPeer.Md5Secret = &networkRef
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for NetworkRef in BgpPeer\n", val)
				}
			case "peer_ip":
				if peerIP, ok := val.(map[string]interface{}); ok {
					bgpPeer.PeerIP = IPAddrIntfToObj(peerIP)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for PeerIP in BgpPeer\n", val)
				}
			case "peer_ip6":
				if peerIP6, ok := val.(map[string]interface{}); ok {
					bgpPeer.PeerIp6 = IPAddrIntfToObj(peerIP6)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for PeerIp6 in BgpPeer\n", val)
				}
			case "remote_as":
				if remoteAs, ok := val.(int32); ok {
					bgpPeer.RemoteAs = &remoteAs
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for RemoteAs in BgpPeer\n", val)
				}
			case "shutdown":
				if shutdown, ok := val.(bool); ok {
					bgpPeer.Shutdown = &shutdown
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for Shutdown in BgpPeer\n", val)
				}
			case "subnet":
				if subnetIntf, ok := val.(map[string]interface{}); ok {
					bgpPeer.Subnet = IAddrPrefixIntfToObj(subnetIntf)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for Subnet in BgpPeer\n", val)
				}
			case "subnet6":
				if subnet6Intf, ok := val.(map[string]interface{}); ok {
					bgpPeer.Subnet6 = IAddrPrefixIntfToObj(subnet6Intf)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for Subnet6 in BgpPeer\n", val)
				}
			default:
				utils.AviLog.Warning.Printf("Unknown key %s in BgpPeer\n", key)
			}
		}
		bgpPeers = append(bgpPeers, &bgpPeer)
	}
	return bgpPeers
}

func BgpProfileIntfToObj(bgpProfileIntf map[string]interface{}) *avimodels.BgpProfile {
	bgpProfile := avimodels.BgpProfile{}
	for key, val := range bgpProfileIntf {
		switch key {
		case "community":
			if community, ok := val.([]string); ok {
				bgpProfile.Community = community
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for Community in BgpProfile\n", val)
			}
		case "hold_time":
			if holdTime, ok := val.(int32); ok {
				bgpProfile.HoldTime = &holdTime
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for HoldTime in BgpProfile\n", val)
			}
		case "ibgp":
			if ibgp, ok := val.(bool); ok {
				bgpProfile.Ibgp = &ibgp
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for Ibgp in BgpProfile\n", val)
			}
		case "ip_communities":
			if ipCommunitiesIntf, ok := val.([]interface{}); ok {
				bgpProfile.IPCommunities = IpCommunitiesIntfToObj(ipCommunitiesIntf)
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for IPCommunities in BgpProfile\n", val)
			}
		case "keep_alive_interval":
			if keepAliveInterval, ok := val.(int32); ok {
				bgpProfile.LocalAs = &keepAliveInterval
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for KeppAliveInterval in BgpProfile\n", val)
			}
		case "local_as":
			if localAs, ok := val.(int32); ok {
				bgpProfile.LocalAs = &localAs
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for LocalAs in BgpProfile\n", val)
			}
		case "peers":
			if bgpPeerIntf, ok := val.([]interface{}); ok {
				bgpProfile.Peers = bgpPeerIntfToObj(bgpPeerIntf)
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for Peers in BgpProfile\n", val)
			}
		case "send_community":
			if sendCommunity, ok := val.(bool); ok {
				bgpProfile.SendCommunity = &sendCommunity
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for SendCommunity in BgpProfile\n", val)
			}
		case "shutdown":
			if shutdown, ok := val.(bool); ok {
				bgpProfile.Shutdown = &shutdown
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for Shutdown in BgpProfile\n", val)
			}
		default:
			utils.AviLog.Warning.Printf("Unknown key %s in BgpProfile\n", key)
		}
	}
	return &bgpProfile
}

func DebugVrfIntfToObj(debugVrfsIntf []interface{}) []*avimodels.DebugVrf {
	debugVrfs := []*avimodels.DebugVrf{}
	for _, dbgVrfIntf := range debugVrfsIntf {
		dbgVrfMap, ok := dbgVrfIntf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("object type %T did not match for DebugVrf\n", dbgVrfIntf)
			continue
		}
		dbgVrf := avimodels.DebugVrf{}
		for key, val := range dbgVrfMap {
			switch key {
			case "flag":
				if flag, ok := val.(string); ok {
					dbgVrf.Flag = &flag
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for flag in DebugVrf\n", val)
				}
			default:
				utils.AviLog.Warning.Printf("Unknown key %s in DebugVrf\n", key)
			}
		}
		debugVrfs = append(debugVrfs, &dbgVrf)

	}
	return debugVrfs
}

func DebugVrfContestIntfToObj(debugVrfContestIntf map[string]interface{}) *avimodels.DebugVrfContext {
	debugVrfContext := avimodels.DebugVrfContext{}
	for key, val := range debugVrfContestIntf {
		switch key {
		case "command_buffer_interval":
			if commandBufferInterval, ok := val.(int32); ok {
				debugVrfContext.CommandBufferInterval = &commandBufferInterval
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for command_buffer_interval in DebugVrfContext\n", val)
			}
		case "command_buffer_size":
			if commandBufferSize, ok := val.(int32); ok {
				debugVrfContext.CommandBufferSize = &commandBufferSize
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for command_buffer_size in DebugVrfContext\n", val)
			}
		case "flags":
			if flags, ok := val.([]interface{}); ok {
				debugVrfContext.Flags = DebugVrfIntfToObj(flags)
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for flags in DebugVrfContext\n", val)
			}
		default:
			utils.AviLog.Warning.Printf("Unknown key %s in DebugVrfContext\n", key)
		}
	}
	return &debugVrfContext
}

func GatewayMonIntfToObj(gatewayMonitersIntf []interface{}) []*avimodels.GatewayMonitor {
	gatewayMonitors := []*avimodels.GatewayMonitor{}
	for _, gatewayMonIntf := range gatewayMonitersIntf {
		gatewayMonMap, ok := gatewayMonIntf.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("object type %T did not match for GatewayMonitor\n", gatewayMonIntf)
			continue
		}
		gatewayMon := avimodels.GatewayMonitor{}
		for key, val := range gatewayMonMap {
			switch key {
			case "gateway_ip":
				if gatewayIP, ok := val.(map[string]interface{}); ok {
					gatewayMon.GatewayIP = IPAddrIntfToObj(gatewayIP)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for gateway_ip in GatewayMonitor\n", val)
				}
			case "gateway_monitor_fail_threashold":
				if gatewayMonitorFailThreashold, ok := val.(int32); ok {
					gatewayMon.GatewayMonitorFailThreshold = &gatewayMonitorFailThreashold
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for gateway_monitor_fail_threashold in GatewayMonitor\n", val)
				}
			case "gateway_monitor_interval":
				if gatewayMonitorInterval, ok := val.(int32); ok {
					gatewayMon.GatewayMonitorInterval = &gatewayMonitorInterval
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for gateway_monitor_interval in GatewayMonitor\n", val)
				}
			case "gateway_monitor_success_threashold":
				if gatewayMonitorSuccessThreshold, ok := val.(int32); ok {
					gatewayMon.GatewayMonitorSuccessThreshold = &gatewayMonitorSuccessThreshold
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for gateway_monitor_success_threashold in GatewayMonitor\n", val)
				}
			case "subnet":
				if subnet, ok := val.(map[string]interface{}); ok {
					gatewayMon.Subnet = IAddrPrefixIntfToObj(subnet)
				} else {
					utils.AviLog.Warning.Printf("wrong object type %T for subnet in GatewayMonitor\n", val)
				}
			default:
				utils.AviLog.Warning.Printf("Unknown key %s in GatewayMonitor\n", key)
			}
		}
		gatewayMonitors = append(gatewayMonitors, &gatewayMon)
	}
	return gatewayMonitors
}

func InternalGatewayMonIntfToObj(InternalGatewayMonIntf map[string]interface{}) *avimodels.InternalGatewayMonitor {
	internalGatewayMonitor := avimodels.InternalGatewayMonitor{}
	for key, val := range InternalGatewayMonIntf {
		switch key {
		case "disable_gateway_monitor":
			if disableGatewayMonitor, ok := val.(bool); ok {
				internalGatewayMonitor.DisableGatewayMonitor = &disableGatewayMonitor
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for disable_gateway_monitor in InternalGatewayMonitor\n", val)
			}
		case "gateway_monitor_failure_threshold":
			if gatewayMonitorFailureThreshold, ok := val.(int32); ok {
				internalGatewayMonitor.GatewayMonitorFailureThreshold = &gatewayMonitorFailureThreshold
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for gateway_monitor_failure_threshold in InternalGatewayMonitor\n", val)
			}
		case "gateway_monitor_interval":
			if gatewayMonitorInterval, ok := val.(int32); ok {
				internalGatewayMonitor.GatewayMonitorInterval = &gatewayMonitorInterval
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for gateway_monitor_interval in InternalGatewayMonitor\n", val)
			}
		case "gateway_monitor_success_threshold":
			if gatewayMonitorSuccessThreshold, ok := val.(int32); ok {
				internalGatewayMonitor.GatewayMonitorSuccessThreshold = &gatewayMonitorSuccessThreshold
			} else {
				utils.AviLog.Warning.Printf("wrong object type %T for gateway_monitor_success_threshold in InternalGatewayMonitor\n", val)
			}
		default:
			utils.AviLog.Warning.Printf("Unknown key %s in InternalGatewayMonitor\n", key)
		}
	}
	return &internalGatewayMonitor
}

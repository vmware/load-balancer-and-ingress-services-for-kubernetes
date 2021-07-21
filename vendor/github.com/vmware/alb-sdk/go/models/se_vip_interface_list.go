// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVipInterfaceList se vip interface list
// swagger:model SeVipInterfaceList
type SeVipInterfaceList struct {

	// Placeholder for description of property is_portchannel of obj type SeVipInterfaceList field type str  type boolean
	IsPortchannel *bool `json:"is_portchannel,omitempty"`

	// List of placement_networks reachable from this interface. Field introduced in 20.1.5.
	Networks []*DiscoveredNetwork `json:"networks,omitempty"`

	// Placeholder for description of property vip_intf_ip of obj type SeVipInterfaceList field type str  type object
	VipIntfIP *IPAddr `json:"vip_intf_ip,omitempty"`

	// Placeholder for description of property vip_intf_ip6 of obj type SeVipInterfaceList field type str  type object
	VipIntfIp6 *IPAddr `json:"vip_intf_ip6,omitempty"`

	// vip_intf_mac of SeVipInterfaceList.
	// Required: true
	VipIntfMac *string `json:"vip_intf_mac"`

	// Number of vlan_id.
	VlanID *int32 `json:"vlan_id,omitempty"`
}

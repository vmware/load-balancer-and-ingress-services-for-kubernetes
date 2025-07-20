// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVipInterfaceList se vip interface list
// swagger:model SeVipInterfaceList
type SeVipInterfaceList struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsPortchannel *bool `json:"is_portchannel,omitempty"`

	// List of placement_networks reachable from this interface. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Networks []*DiscoveredNetwork `json:"networks,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipIntfIP *IPAddr `json:"vip_intf_ip,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipIntfIp6 *IPAddr `json:"vip_intf_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VipIntfMac *string `json:"vip_intf_mac"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VlanID *int32 `json:"vlan_id,omitempty"`
}

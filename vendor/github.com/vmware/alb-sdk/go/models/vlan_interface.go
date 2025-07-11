// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VlanInterface vlan interface
// swagger:model VlanInterface
type VlanInterface struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	// Enable the interface. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IfName *string `json:"if_name"`

	// Enable IPv6 auto configuration. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ip6AutocfgEnabled *bool `json:"ip6_autocfg_enabled,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsMgmt *bool `json:"is_mgmt,omitempty"`

	// VLAN ID. Allowed values are 0-4096. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VlanID *uint32 `json:"vlan_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VnicNetworks []*VNICNetwork `json:"vnic_networks,omitempty"`

	//  It is a reference to an object of type VrfContext. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfRef *string `json:"vrf_ref,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipSummary vip summary
// swagger:model VipSummary
type VipSummary struct {

	// Auto-allocate floating/elastic IP from the Cloud infrastructure. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AutoAllocateFloatingIP *bool `json:"auto_allocate_floating_ip,omitempty"`

	// Auto-allocate VIP from the provided subnet. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AutoAllocateIP *bool `json:"auto_allocate_ip,omitempty"`

	// Specifies whether to auto-allocate only a V4 address, only a V6 address, or one of each type. Enum options - V4_ONLY, V6_ONLY, V4_V6. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AutoAllocateIPType *string `json:"auto_allocate_ip_type,omitempty"`

	// (internal-use) FIP allocated by Avi in the Cloud infrastructure. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AviAllocatedFip *bool `json:"avi_allocated_fip,omitempty"`

	// (internal-use) VIP allocated by Avi in the Cloud infrastructure. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AviAllocatedVip *bool `json:"avi_allocated_vip,omitempty"`

	// Discovered networks providing reachability for client facing Vip IP. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DiscoveredNetworks []*DiscoveredNetwork `json:"discovered_networks,omitempty"`

	// Enable or disable the Vip. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// IPv4 Address of the VIP. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPAddress *IPAddr `json:"ip_address,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumSeAssigned uint32 `json:"num_se_assigned,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumSeRequested uint32 `json:"num_se_requested,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PercentSesUp *int32 `json:"percent_ses_up,omitempty"`

	// Placement networks/subnets to use for vip placement. Field introduced in 22.1.1. Maximum of 10 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PlacementNetworks []*VipPlacementNetwork `json:"placement_networks,omitempty"`

	// Mask applied for the Vip, non-default mask supported only for wildcard Vip. Allowed values are 0-32. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PrefixLength *uint32 `json:"prefix_length,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceEngine []*VipSeAssigned `json:"service_engine,omitempty"`

	// Unique ID associated with the vip. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VipID *string `json:"vip_id,omitempty"`
}

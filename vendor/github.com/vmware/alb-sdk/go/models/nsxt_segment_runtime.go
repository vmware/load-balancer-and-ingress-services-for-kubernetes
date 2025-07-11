// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtSegmentRuntime nsxt segment runtime
// swagger:model NsxtSegmentRuntime
type NsxtSegmentRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Nsxt segment belongs to cloud. It is a reference to an object of type Cloud. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// V6 DHCP ranges configured in Nsxt. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Dhcp6Ranges []string `json:"dhcp6_ranges,omitempty"`

	// IP address management scheme for this Segment associated network. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	// DHCP ranges configured in Nsxt. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DhcpRanges []string `json:"dhcp_ranges,omitempty"`

	// Segment object name. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Network Name. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NwName *string `json:"nw_name,omitempty"`

	// Corresponding network object in Avi. It is a reference to an object of type Network. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NwRef *string `json:"nw_ref,omitempty"`

	// Opaque network Id. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OpaqueNetworkID *string `json:"opaque_network_id,omitempty"`

	// Origin ID applicable to security only cloud. Field introduced in 22.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OriginID *string `json:"origin_id,omitempty"`

	// Nsxt segment belongs to Security only cloud. Field introduced in 22.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SecurityOnlyNsxt *bool `json:"security_only_nsxt,omitempty"`

	// Segment Gateway. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SegmentGw *string `json:"segment_gw,omitempty"`

	// V6 segment Gateway. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SegmentGw6 *string `json:"segment_gw6,omitempty"`

	// Segment Id. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SegmentID *string `json:"segment_id,omitempty"`

	// Segment name. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Segname *string `json:"segname,omitempty"`

	// Segment Cidr. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Subnet *string `json:"subnet,omitempty"`

	// V6 Segment Cidr. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Subnet6 *string `json:"subnet6,omitempty"`

	// Nsxt segment belongs to tenant. It is a reference to an object of type Tenant. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Tier1 router Id. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tier1ID *string `json:"tier1_id,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Uuid. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Segment Vlan ids. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VlanIds []string `json:"vlan_ids,omitempty"`

	// Corresponding vrf context object in Avi. It is a reference to an object of type VrfContext. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`
}

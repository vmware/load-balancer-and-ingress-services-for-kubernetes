// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VrfContext vrf context
// swagger:model VrfContext
type VrfContext struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Key/value vrfcontext attributes. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Attrs []*KeyValue `json:"attrs,omitempty"`

	// BFD configuration profile. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BfdProfile *BfdProfile `json:"bfd_profile,omitempty"`

	// Bgp Local and Peer Info. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BgpProfile *BgpProfile `json:"bgp_profile,omitempty"`

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Configure debug flags for VRF. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Debugvrfcontext *DebugVrfContext `json:"debugvrfcontext,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Configure ping based heartbeat check for gateway in service engines of vrf. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GatewayMon []*GatewayMonitor `json:"gateway_mon,omitempty"`

	// Configure ping based heartbeat check for all default gateways in service engines of vrf. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InternalGatewayMonitor *InternalGatewayMonitor `json:"internal_gateway_monitor,omitempty"`

	// Enable LLDP. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	LldpEnable *bool `json:"lldp_enable,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StaticRoutes []*StaticRoute `json:"static_routes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SystemDefault *bool `json:"system_default,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

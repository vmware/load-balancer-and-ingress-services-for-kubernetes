// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsvipConfig vsvip config
// swagger:model VsvipConfig
type VsvipConfig struct {

	//  It is a reference to an object of type Cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Force placement on all Service Engines in the Service Engine Group (Container clouds only). Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EastWestPlacement *bool `json:"east_west_placement,omitempty"`

	// Name for the VsVip object. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// URL of the VsVip. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URL *string `json:"url,omitempty"`

	// UUID of the VsVip object. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// List of Virtual Service IPs and other shareable entities. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Vip []*Vip `json:"vip,omitempty"`

	// List of virtual services the pool is assigned to. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Virtualservices []*VsRefs `json:"virtualservices,omitempty"`

	// Virtual Routing Context that the Virtual Service is bound to. This is used to provide the isolation of the set of networks the application is attached to. It is a reference to an object of type VrfContext. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`
}

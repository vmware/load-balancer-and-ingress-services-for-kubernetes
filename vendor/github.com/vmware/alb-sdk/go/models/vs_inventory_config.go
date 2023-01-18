// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsInventoryConfig vs inventory config
// swagger:model VsInventoryConfig
type VsInventoryConfig struct {

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Service discovery specific data including fully qualified domain name, type and Time-To-Live of the DNS record. Note that only one of fqdn and dns_info setting is allowed. Maximum of 1000 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSInfo []*DNSInfo `json:"dns_info,omitempty"`

	// Force placement on all SE's in service group (Mesos mode only). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EastWestPlacement *bool `json:"east_west_placement,omitempty"`

	// Enable or disable the Virtual Service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// DNS resolvable, fully qualified domain name of the virtualservice. Only one of 'fqdn' and 'dns_info' configuration is allowed. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Fqdn *string `json:"fqdn,omitempty"`

	// Name for the Virtual Service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// The pool is an object that contains destination servers and related attributes such as load-balancing and persistence. It is a reference to an object of type Pool. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolRef *string `json:"pool_ref,omitempty"`

	// The Service Engine Group to use for this Virtual Service. Moving to a new SE Group is disruptive to existing connections for this VS. It is a reference to an object of type ServiceEngineGroup. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	// List of Services defined for this Virtual Service. Field introduced in 22.1.1. Maximum of 2048 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Services []*Service `json:"services,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Specify if this is a normal Virtual Service, or if it is the parent or child of an SNI-enabled virtual hosted Virtual Service. Enum options - VS_TYPE_NORMAL, VS_TYPE_VH_PARENT, VS_TYPE_VH_CHILD. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	// URL of the Virtual Service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URL *string `json:"url,omitempty"`

	// UUID of the Virtual Service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// The exact name requested from the client's SNI-enabled TLS hello domain name field. If this is a match, the parent VS will forward the connection to this child VS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VhDomainName []string `json:"vh_domain_name,omitempty"`

	// List of Virtual Service IPs. While creating a 'Shared VS',please use vsvip_ref to point to the shared entities. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Vip []*Vip `json:"vip,omitempty"`

	// Virtual Routing Context that the Virtual Service is bound to. This is used to provide the isolation of the set of networks the application is attached to. It is a reference to an object of type VrfContext. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`

	// Mostly used during the creation of Shared VS, this field refers to entities that can be shared across Virtual Services. It is a reference to an object of type VsVip. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsvipRef *string `json:"vsvip_ref,omitempty"`
}

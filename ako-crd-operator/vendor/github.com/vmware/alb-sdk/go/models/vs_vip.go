// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsVip vs vip
// swagger:model VsVip
type VsVip struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// LOCAL_PREF to be used for VsVip advertised. Applicable only over iBGP. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BgpLocalPreference uint32 `json:"bgp_local_preference,omitempty"`

	// Number of times the local AS should be prepended additionally to VsVip. Applicable only over eBGP. Allowed values are 1-10. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BgpNumAsPathPrepend uint32 `json:"bgp_num_as_path_prepend,omitempty"`

	// Select BGP peers, using peer label, for VsVip advertisement. Field introduced in 20.1.5. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BgpPeerLabels []string `json:"bgp_peer_labels,omitempty"`

	//  It is a reference to an object of type Cloud. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Service discovery specific data including fully qualified domain name, type and Time-To-Live of the DNS record. This takes effect only if dns profile isassociated with cloud. Field introduced in 17.1.1. Maximum of 1000 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSInfo []*DNSInfo `json:"dns_info,omitempty"`

	// Force placement on all Service Engines in the Service Engine Group (Container clouds only). Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EastWestPlacement *bool `json:"east_west_placement,omitempty"`

	// Determines the set of IPAM networks to use for this VsVip. Selector type must be SELECTOR_IPAM and only one label is supported. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPAMSelector *Selector `json:"ipam_selector,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Name for the VsVip object. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// This sets the placement scope of virtualservice to given tier1 logical router in Nsx-t. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tier1Lr *string `json:"tier1_lr,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// This overrides the cloud level default and needs to match the SE Group value in which it will be used if the SE Group use_standard_alb value is set. This is only used when FIP is used for VS on Azure Cloud. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UseStandardAlb *bool `json:"use_standard_alb,omitempty"`

	// UUID of the VsVip object. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// List of Virtual Service IPs and other shareable entities. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vip []*Vip `json:"vip,omitempty"`

	// Virtual Routing Context that the Virtual Service is bound to. This is used to provide the isolation of the set of networks the application is attached to. It is a reference to an object of type VrfContext. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`

	// Checksum of cloud configuration for VsVip. Internally set by cloud connector. Field introduced in 17.2.9, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsvipCloudConfigCksum *string `json:"vsvip_cloud_config_cksum,omitempty"`
}

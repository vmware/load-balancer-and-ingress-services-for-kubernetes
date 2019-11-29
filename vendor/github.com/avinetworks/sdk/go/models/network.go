package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Network network
// swagger:model Network
type Network struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Placeholder for description of property configured_subnets of obj type Network field type str  type object
	ConfiguredSubnets []*Subnet `json:"configured_subnets,omitempty"`

	// Select the IP address management scheme for this Network.
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	// When selected, excludes all discovered subnets in this network from consideration for virtual service placement.
	ExcludeDiscoveredSubnets *bool `json:"exclude_discovered_subnets,omitempty"`

	// Enable IPv6 auto configuration. Field introduced in 18.1.1.
	Ip6AutocfgEnabled *bool `json:"ip6_autocfg_enabled,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property synced_from_se of obj type Network field type str  type boolean
	SyncedFromSe *bool `json:"synced_from_se,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Placeholder for description of property vcenter_dvs of obj type Network field type str  type boolean
	VcenterDvs *bool `json:"vcenter_dvs,omitempty"`

	//  It is a reference to an object of type VIMgrNWRuntime.
	// Read Only: true
	VimgrnwRef *string `json:"vimgrnw_ref,omitempty"`

	//  It is a reference to an object of type VrfContext.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`
}

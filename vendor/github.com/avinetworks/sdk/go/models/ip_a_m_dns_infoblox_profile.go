package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSInfobloxProfile ipam Dns infoblox profile
// swagger:model IpamDnsInfobloxProfile
type IPAMDNSInfobloxProfile struct {

	// DNS view.
	DNSView *string `json:"dns_view,omitempty"`

	// Custom parameters that will passed to the Infoblox provider as extensible attributes. Field introduced in 18.2.7.
	ExtensibleAttributes []*CustomParams `json:"extensible_attributes,omitempty"`

	// Address of Infoblox appliance.
	// Required: true
	IPAddress *IPAddr `json:"ip_address"`

	// Network view.
	NetworkView *string `json:"network_view,omitempty"`

	// Password for API access for Infoblox appliance.
	// Required: true
	Password *string `json:"password"`

	// Subnets to use for Infoblox IP allocation. Field introduced in 18.2.8.
	UsableAllocSubnets []*InfobloxSubnet `json:"usable_alloc_subnets,omitempty"`

	// Usable domains to pick from Infoblox.
	UsableDomains []string `json:"usable_domains,omitempty"`

	// This field is deprecated, use usable_alloc_subnets instead. Field deprecated in 18.2.8.
	UsableSubnets []*IPAddrPrefix `json:"usable_subnets,omitempty"`

	// Username for API access for Infoblox appliance.
	// Required: true
	Username *string `json:"username"`

	// WAPI version.
	WapiVersion *string `json:"wapi_version,omitempty"`
}

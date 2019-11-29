package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSInfobloxProfile ipam Dns infoblox profile
// swagger:model IpamDnsInfobloxProfile
type IPAMDNSInfobloxProfile struct {

	// DNS view.
	DNSView *string `json:"dns_view,omitempty"`

	// Address of Infoblox appliance.
	// Required: true
	IPAddress *IPAddr `json:"ip_address"`

	// Network view.
	NetworkView *string `json:"network_view,omitempty"`

	// Password for API access for Infoblox appliance.
	// Required: true
	Password *string `json:"password"`

	// Usable domains to pick from Infoblox.
	UsableDomains []string `json:"usable_domains,omitempty"`

	// Usable subnets to pick from Infoblox.
	UsableSubnets []*IPAddrPrefix `json:"usable_subnets,omitempty"`

	// Username for API access for Infoblox appliance.
	// Required: true
	Username *string `json:"username"`

	// WAPI version.
	WapiVersion *string `json:"wapi_version,omitempty"`
}

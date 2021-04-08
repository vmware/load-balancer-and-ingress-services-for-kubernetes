package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSCustomProfile ipam Dns custom profile
// swagger:model IpamDnsCustomProfile
type IPAMDNSCustomProfile struct {

	//  It is a reference to an object of type CustomIpamDnsProfile. Field introduced in 17.1.1.
	CustomIPAMDNSProfileRef *string `json:"custom_ipam_dns_profile_ref,omitempty"`

	// Custom parameters that will passed to the IPAM/DNS provider including but not limited to provider credentials and API version. Field introduced in 17.1.1.
	DynamicParams []*CustomParams `json:"dynamic_params,omitempty"`

	// Usable domains. Field introduced in 17.2.2.
	UsableDomains []string `json:"usable_domains,omitempty"`

	// Usable subnets. Field introduced in 17.2.2.
	UsableSubnets []*IPAddrPrefix `json:"usable_subnets,omitempty"`
}

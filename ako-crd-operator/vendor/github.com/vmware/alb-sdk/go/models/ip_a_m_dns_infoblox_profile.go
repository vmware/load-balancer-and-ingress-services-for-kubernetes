// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSInfobloxProfile ipam Dns infoblox profile
// swagger:model IpamDnsInfobloxProfile
type IPAMDNSInfobloxProfile struct {

	// DNS view used for Infoblox host record creation, If this field is not configured by the user, then its value will be set to 'default'. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSView *string `json:"dns_view,omitempty"`

	// Custom parameters that will passed to the Infoblox provider as extensible attributes. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExtensibleAttributes []*CustomParams `json:"extensible_attributes,omitempty"`

	// IPv6 Address of Infoblox appliance. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ip6Address *IPAddr `json:"ip6_address,omitempty"`

	// IPv4 Address of Infoblox appliance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAddress *IPAddr `json:"ip_address,omitempty"`

	// Network view used for Infoblox host record creation, If this field is not configured by the user, then its value will be set to 'default'. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkView *string `json:"network_view,omitempty"`

	// Password for API access for Infoblox appliance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Password *string `json:"password"`

	// url of the profile writen to /etc/hosts for HA between IPv6 and IPv4. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ProfileURL *string `json:"profile_url,omitempty"`

	// Subnets to use for Infoblox IP allocation. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UsableAllocSubnets []*InfobloxSubnet `json:"usable_alloc_subnets,omitempty"`

	// Usable domains to pick from Infoblox. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UsableDomains []string `json:"usable_domains,omitempty"`

	// Username for API access for Infoblox appliance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Username *string `json:"username"`

	// WAPI version. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WapiVersion *string `json:"wapi_version,omitempty"`
}

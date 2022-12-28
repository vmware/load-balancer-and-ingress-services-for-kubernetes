// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSInfobloxProfile ipam Dns infoblox profile
// swagger:model IpamDnsInfobloxProfile
type IPAMDNSInfobloxProfile struct {

	// DNS view. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSView *string `json:"dns_view,omitempty"`

	// Custom parameters that will passed to the Infoblox provider as extensible attributes. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExtensibleAttributes []*CustomParams `json:"extensible_attributes,omitempty"`

	// Address of Infoblox appliance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IPAddress *IPAddr `json:"ip_address"`

	// Network view. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkView *string `json:"network_view,omitempty"`

	// Password for API access for Infoblox appliance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Password *string `json:"password"`

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

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSCustomProfile ipam Dns custom profile
// swagger:model IpamDnsCustomProfile
type IPAMDNSCustomProfile struct {

	//  It is a reference to an object of type CustomIpamDnsProfile. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CustomIPAMDNSProfileRef *string `json:"custom_ipam_dns_profile_ref,omitempty"`

	// Custom parameters that will passed to the IPAM/DNS provider including but not limited to provider credentials and API version. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DynamicParams []*CustomParams `json:"dynamic_params,omitempty"`

	// Networks or Subnets to use for Custom IPAM IP allocation. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UsableAllocSubnets []*CustomIPAMSubnet `json:"usable_alloc_subnets,omitempty"`

	// Usable domains. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UsableDomains []string `json:"usable_domains,omitempty"`
}

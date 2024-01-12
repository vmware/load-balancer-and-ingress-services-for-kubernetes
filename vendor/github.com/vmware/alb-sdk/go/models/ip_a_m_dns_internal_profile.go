// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSInternalProfile ipam Dns internal profile
// swagger:model IpamDnsInternalProfile
type IPAMDNSInternalProfile struct {

	// List of service domains. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSServiceDomain []*DNSServiceDomain `json:"dns_service_domain,omitempty"`

	// Avi VirtualService to be used for serving DNS records. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSVirtualserviceRef *string `json:"dns_virtualservice_ref,omitempty"`

	// Default TTL for all records, overridden by TTL value for each service domain configured in DnsServiceDomain. Allowed values are 1-604800. Unit is SEC. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 30), Basic edition(Allowed values- 30), Enterprise with Cloud Services edition.
	TTL *uint32 `json:"ttl,omitempty"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	UsableNetworks []*IPAMUsableNetwork `json:"usable_networks,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSInternalProfile ipam Dns internal profile
// swagger:model IpamDnsInternalProfile
type IPAMDNSInternalProfile struct {

	// List of service domains.
	DNSServiceDomain []*DNSServiceDomain `json:"dns_service_domain,omitempty"`

	// Avi VirtualService to be used for serving DNS records. It is a reference to an object of type VirtualService.
	DNSVirtualserviceRef *string `json:"dns_virtualservice_ref,omitempty"`

	// Default TTL for all records, overridden by TTL value for each service domain configured in DnsServiceDomain. Allowed values are 1-604800.
	TTL *int32 `json:"ttl,omitempty"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. It is a reference to an object of type Network.
	UsableNetworkRefs []string `json:"usable_network_refs,omitempty"`
}

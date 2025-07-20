// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSServiceApplicationProfile Dns service application profile
// swagger:model DnsServiceApplicationProfile
type DNSServiceApplicationProfile struct {

	// Respond to AAAA queries with empty response when there are only IPV4 records. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AaaaEmptyResponse *bool `json:"aaaa_empty_response,omitempty"`

	// Email address of the administrator responsible for this zone . This field is used in SOA records (rname) pertaining to all domain names specified as authoritative domain names. If not configured, the default value 'hostmaster' is used in SOA responses. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdminEmail *string `json:"admin_email,omitempty"`

	// The maximum time allowed for a client to transmit an entire DNS request over TCP. This helps mitigate various forms of SlowLoris attacks. Allowed values are 10-100000000. Field introduced in 22.1.5, 30.1.2, 30.2.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientDNSTCPRequestTimeout *uint32 `json:"client_dns_tcp_request_timeout,omitempty"`

	// If enabled, the Service Engine initiates closure of client TCP connections after the first DNS response, for pass-through/proxy cases. This behavior applies to all DNS request types other than AX-FR. Field introduced in 21.1.7, 22.1.4, 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloseTCPConnectionPostResponse *bool `json:"close_tcp_connection_post_response,omitempty"`

	// Enable DNS query/response over TCP. This enables analytics for pass-through queries as well. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSOverTCPEnabled *bool `json:"dns_over_tcp_enabled,omitempty"`

	// DNS zones hosted on this Virtual Service. Field introduced in 18.2.6. Maximum of 100 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSZones []*DNSZone `json:"dns_zones,omitempty"`

	// Subdomain names serviced by this Virtual Service. These are configured as Ends-With semantics. Maximum of 100 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DomainNames []string `json:"domain_names,omitempty"`

	// Enable stripping of EDNS client subnet (ecs) option towards client if DNS service inserts ecs option in the DNS query towards upstream servers. Field introduced in 17.1.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EcsStrippingEnabled *bool `json:"ecs_stripping_enabled,omitempty"`

	// Enable DNS service to be aware of EDNS (Extension mechanism for DNS). EDNS extensions are parsed and shown in logs. For GSLB services, the EDNS client subnet option can be used to influence Load Balancing. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Edns *bool `json:"edns,omitempty"`

	// Specifies the IP address prefix length to use in the EDNS client subnet (ECS) option. When the incoming request does not have any ECS option and the prefix length is specified, an ECS option is inserted in the request passed to upstream server. If the incoming request already has an ECS option, the prefix length (and correspondingly the address) in the ECS option is updated, with the minimum of the prefix length present in the incoming and the configured prefix length, before passing the request to upstream server. Allowed values are 1-32. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EdnsClientSubnetPrefixLen *uint32 `json:"edns_client_subnet_prefix_len,omitempty"`

	// Drop or respond to client when the DNS service encounters an error processing a client query. By default, such a request is dropped without any response, or passed through to a passthrough pool, if configured. When set to respond, an appropriate response is sent to client, e.g. NXDOMAIN response for non-existent records, empty NOERROR response for unsupported queries, etc. Enum options - DNS_ERROR_RESPONSE_ERROR, DNS_ERROR_RESPONSE_NONE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErrorResponse *string `json:"error_response,omitempty"`

	// The <domain-name>  of the name server that was the original or primary source of data for this zone. This field is used in SOA records (mname) pertaining to all domain names specified as authoritative domain names. If not configured, domain name is used as name server in SOA response. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NameServer *string `json:"name_server,omitempty"`

	// Specifies the TTL value (in seconds) for SOA (Start of Authority) (corresponding to a authoritative domain owned by this DNS Virtual Service) record's minimum TTL served by the DNS Virtual Service. Allowed values are 0-86400. Field introduced in 17.2.4. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NegativeCachingTTL *uint32 `json:"negative_caching_ttl,omitempty"`

	// Specifies the number of IP addresses returned by the DNS Service. Enter 0 to return all IP addresses. Allowed values are 1-20. Special values are 0- Return all IP addresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumDNSIP *uint32 `json:"num_dns_ip,omitempty"`

	// Specifies the TTL value (in seconds) for records served by DNS Service. Allowed values are 0-86400. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TTL *uint32 `json:"ttl,omitempty"`
}

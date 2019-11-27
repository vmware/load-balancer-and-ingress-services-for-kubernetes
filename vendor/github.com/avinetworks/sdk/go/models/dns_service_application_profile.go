package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSServiceApplicationProfile Dns service application profile
// swagger:model DnsServiceApplicationProfile
type DNSServiceApplicationProfile struct {

	// Respond to AAAA queries with empty response when there are only IPV4 records.
	AaaaEmptyResponse *bool `json:"aaaa_empty_response,omitempty"`

	// Domain names authoritatively serviced by this Virtual Service. These are configured as Ends-With semantics. Queries for FQDNs that are subdomains of this domain and do not have any DNS record in Avi are dropped or NXDomain response sent. . Field introduced in 17.1.6,17.2.2.
	AuthoritativeDomainNames []string `json:"authoritative_domain_names,omitempty"`

	// Enable DNS query/response over TCP. This enables analytics for pass-through queries as well. Field introduced in 17.1.1.
	DNSOverTCPEnabled *bool `json:"dns_over_tcp_enabled,omitempty"`

	// Subdomain names serviced by this Virtual Service. These are configured as Ends-With semantics.
	DomainNames []string `json:"domain_names,omitempty"`

	// Enable stripping of EDNS client subnet (ecs) option towards client if DNS service inserts ecs option in the DNS query towards upstream servers. Field introduced in 17.1.5.
	EcsStrippingEnabled *bool `json:"ecs_stripping_enabled,omitempty"`

	// Enable DNS service to be aware of EDNS (Extension mechanism for DNS). EDNS extensions are parsed and shown in logs. For GSLB services, the EDNS client subnet option can be used to influence Load Balancing. Field introduced in 17.1.1.
	Edns *bool `json:"edns,omitempty"`

	// Specifies the IP address prefix length to use in the EDNS client subnet (ECS) option. When the incoming request does not have any ECS option and the prefix length is specified, an ECS option is inserted in the request passed to upstream server. If the incoming request already has an ECS option, the prefix length (and correspondingly the address) in the ECS option is updated, with the minimum of the prefix length present in the incoming and the configured prefix length, before passing the request to upstream server. Allowed values are 1-32. Field introduced in 17.1.3.
	EdnsClientSubnetPrefixLen *int32 `json:"edns_client_subnet_prefix_len,omitempty"`

	// Drop or respond to client when the DNS service encounters an error processing a client query. By default, such a request is dropped without any response, or passed through to a passthrough pool, if configured. When set to respond, an appropriate response is sent to client, e.g. NXDOMAIN response for non-existent records, empty NOERROR response for unsupported queries, etc. Enum options - DNS_ERROR_RESPONSE_ERROR, DNS_ERROR_RESPONSE_NONE.
	ErrorResponse *string `json:"error_response,omitempty"`

	// Specifies the TTL value (in seconds) for SOA (Start of Authority) (corresponding to a authoritative domain owned by this DNS Virtual Service) record's minimum TTL served by the DNS Virtual Service. Allowed values are 0-86400. Field introduced in 17.2.4.
	NegativeCachingTTL *int32 `json:"negative_caching_ttl,omitempty"`

	// Specifies the number of IP addresses returned by the DNS Service. Enter 0 to return all IP addresses. Allowed values are 1-20. Special values are 0- 'Return all IP addresses'.
	NumDNSIP *int32 `json:"num_dns_ip,omitempty"`

	// Specifies the TTL value (in seconds) for records served by DNS Service. Allowed values are 0-86400.
	TTL *int32 `json:"ttl,omitempty"`
}

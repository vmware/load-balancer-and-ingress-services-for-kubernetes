package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSServiceDomain Dns service domain
// swagger:model DnsServiceDomain
type DNSServiceDomain struct {

	// Service domain *string used for FQDN.
	// Required: true
	DomainName *string `json:"domain_name"`

	// Specifies the number of A records returned by Avi DNS Service. Allowed values are 0-20. Special values are 0- 'Return all IP addresses'.
	NumDNSIP *int32 `json:"num_dns_ip,omitempty"`

	// Third-party Authoritative domain requests are delegated toDNS VirtualService's pool of nameservers.
	PassThrough *bool `json:"pass_through,omitempty"`

	// TTL value for DNS records. Allowed values are 1-604800.
	RecordTTL *int32 `json:"record_ttl,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSServiceDomain Dns service domain
// swagger:model DnsServiceDomain
type DNSServiceDomain struct {

	// Service domain *string used for FQDN.
	// Required: true
	DomainName *string `json:"domain_name"`

	// [DEPRECATED] Useless fieldPlease refer to DnsServiceApplicationProfile's num_dns_ip for default valuePlease refer to VsVip's dns_info num_records_in_response for user config valueSpecifies the number of A recordsreturned by Avi DNS Service. Field deprecated in 20.1.5.
	NumDNSIP *int32 `json:"num_dns_ip,omitempty"`

	// Third-party Authoritative domain requests are delegated toDNS VirtualService's pool of nameservers.
	PassThrough *bool `json:"pass_through,omitempty"`

	// TTL value for DNS records. Allowed values are 1-604800. Unit is SEC.
	RecordTTL *int32 `json:"record_ttl,omitempty"`
}

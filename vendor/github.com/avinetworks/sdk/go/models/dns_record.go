package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRecord Dns record
// swagger:model DnsRecord
type DNSRecord struct {

	// Specifies the algorithm to pick the IP address(es) to be returned, when multiple entries are configured. This does not apply if num_records_in_response is 0. Default is round-robin. Enum options - DNS_RECORD_RESPONSE_ROUND_ROBIN, DNS_RECORD_RESPONSE_CONSISTENT_HASH. Field introduced in 17.1.1.
	Algorithm *string `json:"algorithm,omitempty"`

	// Canonical name in CNAME record.
	Cname *DNSCnameRdata `json:"cname,omitempty"`

	// Configured FQDNs are delegated domains (i.e. they represent a zone cut). Field introduced in 17.1.2.
	Delegated *bool `json:"delegated,omitempty"`

	// Details of DNS record.
	Description *string `json:"description,omitempty"`

	// Fully Qualified Domain Name.
	Fqdn []string `json:"fqdn,omitempty"`

	// IPv6 address in AAAA record. Field introduced in 18.1.1.
	Ip6Address []*DNSAAAARdata `json:"ip6_address,omitempty"`

	// IP address in A record.
	IPAddress []*DNSARdata `json:"ip_address,omitempty"`

	// Name Server information in NS record. Field introduced in 17.1.1.
	Ns []*DNSNsRdata `json:"ns,omitempty"`

	// Specifies the number of records returned by the DNS service. Enter 0 to return all records. Default is 0. Allowed values are 0-20. Special values are 0- 'Return all records'. Field introduced in 17.1.1.
	NumRecordsInResponse *int32 `json:"num_records_in_response,omitempty"`

	// Service locator info in SRV record.
	ServiceLocator []*DNSSrvRdata `json:"service_locator,omitempty"`

	// Time To Live for this DNS record.
	TTL *int32 `json:"ttl,omitempty"`

	// DNS record type. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY.
	// Required: true
	Type *string `json:"type"`

	// Enable wild-card match of fqdn  if an exact match is not found in the DNS table, the longest match is chosen by wild-carding the fqdn in the DNS request. Default is false. Field introduced in 17.1.1.
	WildcardMatch *bool `json:"wildcard_match,omitempty"`
}

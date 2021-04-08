package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSInfo Dns info
// swagger:model DnsInfo
type DNSInfo struct {

	// Specifies the algorithm to pick the IP address(es) to be returned, when multiple entries are configured. This does not apply if num_records_in_response is 0. Default is consistent hash. Enum options - DNS_RECORD_RESPONSE_ROUND_ROBIN, DNS_RECORD_RESPONSE_CONSISTENT_HASH. Field introduced in 17.1.1.
	Algorithm *string `json:"algorithm,omitempty"`

	// Canonical name in CNAME record. Field introduced in 17.2.1.
	Cname *DNSCnameRdata `json:"cname,omitempty"`

	// Fully qualified domain name.
	Fqdn *string `json:"fqdn,omitempty"`

	// Any metadata associated with this record. Field introduced in 17.2.2.
	// Read Only: true
	Metadata *string `json:"metadata,omitempty"`

	// Specifies the number of records returned for this FQDN. Enter 0 to return all records. Default is 0. Allowed values are 0-20. Special values are 0- 'Return all records'. Field introduced in 17.1.1.
	NumRecordsInResponse *int32 `json:"num_records_in_response,omitempty"`

	// Time to live for fqdn record. Default value is chosen from DNS profile for this cloud if no value provided.
	TTL *int32 `json:"ttl,omitempty"`

	// DNS record type. Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY.
	Type *string `json:"type,omitempty"`
}

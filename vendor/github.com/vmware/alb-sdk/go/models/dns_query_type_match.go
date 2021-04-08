package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSQueryTypeMatch Dns query type match
// swagger:model DnsQueryTypeMatch
type DNSQueryTypeMatch struct {

	// Criterion to use for matching the DNS query typein the question section. Enum options - IS_IN, IS_NOT_IN. Field introduced in 17.1.1.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// DNS query types in the request query . Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY. Field introduced in 17.1.1.
	QueryType []string `json:"query_type,omitempty"`
}

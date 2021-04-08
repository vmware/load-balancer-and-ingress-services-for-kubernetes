package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSTransportProtocolMatch Dns transport protocol match
// swagger:model DnsTransportProtocolMatch
type DNSTransportProtocolMatch struct {

	// Criterion to use for matching the DNS transport protocol. Enum options - IS_IN, IS_NOT_IN. Field introduced in 17.1.1.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Protocol to match against transport protocol used by DNS query. Enum options - DNS_OVER_UDP, DNS_OVER_TCP. Field introduced in 17.1.1.
	// Required: true
	Protocol *string `json:"protocol"`
}

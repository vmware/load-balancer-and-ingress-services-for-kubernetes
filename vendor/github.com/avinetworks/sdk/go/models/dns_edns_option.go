package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSEdnsOption Dns edns option
// swagger:model DnsEdnsOption
type DNSEdnsOption struct {

	// Address family. Field introduced in 17.1.1.
	AddrFamily *int32 `json:"addr_family,omitempty"`

	// EDNS option code. Enum options - EDNS_OPTION_CODE_NSID, EDNS_OPTION_CODE_DNSSEC_DAU, EDNS_OPTION_CODE_DNSSEC_DHU, EDNS_OPTION_CODE_DNSSEC_N3U, EDNS_OPTION_CODE_CLIENT_SUBNET, EDNS_OPTION_CODE_EXPIRE, EDNS_OPTION_CODE_COOKIE, EDNS_OPTION_CODE_TCP_KEEPALIVE, EDNS_OPTION_CODE_PADDING, EDNS_OPTION_CODE_CHAIN. Field introduced in 17.1.1.
	// Required: true
	Code *string `json:"code"`

	// Scope prefix length of address. Field introduced in 17.1.1.
	ScopePrefixLen *int32 `json:"scope_prefix_len,omitempty"`

	// Source prefix length of address. Field introduced in 17.1.1.
	SourcePrefixLen *int32 `json:"source_prefix_len,omitempty"`

	// IPv4 address of the client subnet. Field introduced in 17.1.1.
	SubnetIP *int32 `json:"subnet_ip,omitempty"`
}

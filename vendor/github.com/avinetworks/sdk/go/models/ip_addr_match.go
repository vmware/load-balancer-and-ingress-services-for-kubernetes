package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAddrMatch Ip addr match
// swagger:model IpAddrMatch
type IPAddrMatch struct {

	// IP address(es).
	Addrs []*IPAddr `json:"addrs,omitempty"`

	// UUID of IP address group(s). It is a reference to an object of type IpAddrGroup.
	GroupRefs []string `json:"group_refs,omitempty"`

	// Criterion to use for IP address matching the HTTP request. Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// IP address prefix(es).
	Prefixes []*IPAddrPrefix `json:"prefixes,omitempty"`

	// IP address range(s).
	Ranges []*IPAddrRange `json:"ranges,omitempty"`
}

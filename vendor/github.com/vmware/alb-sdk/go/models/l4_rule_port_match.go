package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4RulePortMatch l4 rule port match
// swagger:model L4RulePortMatch
type L4RulePortMatch struct {

	// Criterion to use for Virtual Service port matching. Enum options - IS_IN, IS_NOT_IN. Field introduced in 17.2.7.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Range of TCP/UDP port numbers of the Virtual Service. Field introduced in 17.2.7.
	PortRanges []*PortRange `json:"port_ranges,omitempty"`

	// Virtual Service's listening port(s). Allowed values are 1-65535. Field introduced in 17.2.7.
	Ports []int64 `json:"ports,omitempty,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4RuleProtocolMatch l4 rule protocol match
// swagger:model L4RuleProtocolMatch
type L4RuleProtocolMatch struct {

	// Criterion to use for transport protocol matching. Enum options - IS_IN, IS_NOT_IN. Field introduced in 17.2.7.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Transport protocol to match. Enum options - PROTOCOL_ICMP, PROTOCOL_TCP, PROTOCOL_UDP. Field introduced in 17.2.7.
	// Required: true
	Protocol *string `json:"protocol"`
}

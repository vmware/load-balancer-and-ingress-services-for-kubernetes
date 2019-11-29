package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafExclusionType waf exclusion type
// swagger:model WafExclusionType
type WafExclusionType struct {

	// Case sensitivity to use for the matching. Enum options - SENSITIVE, INSENSITIVE. Field introduced in 17.2.8.
	// Required: true
	MatchCase *string `json:"match_case"`

	// String Operation to use for matching the Exclusion. Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH. Field introduced in 17.2.8.
	// Required: true
	MatchOp *string `json:"match_op"`
}

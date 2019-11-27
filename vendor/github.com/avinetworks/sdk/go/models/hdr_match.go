package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HdrMatch hdr match
// swagger:model HdrMatch
type HdrMatch struct {

	// Name of the HTTP header whose value is to be matched.
	// Required: true
	Hdr *string `json:"hdr"`

	// Case sensitivity to use for the match. Enum options - SENSITIVE, INSENSITIVE.
	MatchCase *string `json:"match_case,omitempty"`

	// Criterion to use for matching headers in the HTTP request. Enum options - HDR_EXISTS, HDR_DOES_NOT_EXIST, HDR_BEGINS_WITH, HDR_DOES_NOT_BEGIN_WITH, HDR_CONTAINS, HDR_DOES_NOT_CONTAIN, HDR_ENDS_WITH, HDR_DOES_NOT_END_WITH, HDR_EQUALS, HDR_DOES_NOT_EQUAL.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// String values to match in the HTTP header.
	Value []string `json:"value,omitempty"`
}

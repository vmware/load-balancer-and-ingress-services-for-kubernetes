package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PathMatch path match
// swagger:model PathMatch
type PathMatch struct {

	// Case sensitivity to use for the matching. Enum options - SENSITIVE, INSENSITIVE.
	MatchCase *string `json:"match_case,omitempty"`

	// Criterion to use for matching the path in the HTTP request URI. Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// String values.
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the *string group(s). It is a reference to an object of type StringGroup.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StringMatch *string match
// swagger:model StringMatch
type StringMatch struct {

	// Criterion to use for *string matching the HTTP request. Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// String value(s).
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the *string group(s). It is a reference to an object of type StringGroup.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

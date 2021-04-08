package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// QueryMatch query match
// swagger:model QueryMatch
type QueryMatch struct {

	// Case sensitivity to use for the match. Enum options - SENSITIVE, INSENSITIVE.
	MatchCase *string `json:"match_case,omitempty"`

	// Criterion to use for matching the query in HTTP request URI. Enum options - QUERY_MATCH_CONTAINS.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// String value(s).
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the *string group(s). It is a reference to an object of type StringGroup.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

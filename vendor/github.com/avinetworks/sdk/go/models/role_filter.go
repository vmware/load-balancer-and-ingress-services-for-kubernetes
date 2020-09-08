package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RoleFilter role filter
// swagger:model RoleFilter
type RoleFilter struct {

	// Enable this filter. Field introduced in 20.1.3.
	Enabled *bool `json:"enabled,omitempty"`

	// Label key to match against objects for values. Field introduced in 20.1.3.
	// Required: true
	MatchLabel *RoleFilterMatchLabel `json:"match_label"`

	// Label match operation criteria. Enum options - ROLE_FILTER_EQUALS, ROLE_FILTER_DOES_NOT_EQUAL, ROLE_FILTER_GLOB_MATCH, ROLE_FILTER_GLOB_DOES_NOT_MATCH. Field introduced in 20.1.3.
	MatchOperation *string `json:"match_operation,omitempty"`

	// Name for the filter. Field introduced in 20.1.3.
	Name *string `json:"name,omitempty"`
}

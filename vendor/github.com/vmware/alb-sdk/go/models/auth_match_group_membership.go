package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthMatchGroupMembership auth match group membership
// swagger:model AuthMatchGroupMembership
type AuthMatchGroupMembership struct {

	// rule match criteria. Enum options - AUTH_MATCH_CONTAINS, AUTH_MATCH_DOES_NOT_CONTAIN, AUTH_MATCH_REGEX.
	// Required: true
	Criteria *string `json:"criteria"`

	// groups of AuthMatchGroupMembership.
	Groups []string `json:"groups,omitempty"`
}

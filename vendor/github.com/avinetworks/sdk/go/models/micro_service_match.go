package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MicroServiceMatch micro service match
// swagger:model MicroServiceMatch
type MicroServiceMatch struct {

	// UUID of Micro Service group(s). It is a reference to an object of type MicroServiceGroup.
	// Required: true
	GroupRef *string `json:"group_ref"`

	// Criterion to use for Micro Service matching the HTTP request. Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`
}

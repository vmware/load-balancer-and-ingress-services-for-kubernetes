package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PortMatch port match
// swagger:model PortMatch
type PortMatch struct {

	// Criterion to use for port matching the HTTP request. Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Listening TCP port(s). Allowed values are 1-65535.
	Ports []int64 `json:"ports,omitempty,omitempty"`
}

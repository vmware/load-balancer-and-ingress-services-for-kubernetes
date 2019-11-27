package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ProtocolMatch protocol match
// swagger:model ProtocolMatch
type ProtocolMatch struct {

	// Criterion to use for protocol matching the HTTP request. Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// HTTP or HTTPS protocol. Enum options - HTTP, HTTPS.
	// Required: true
	Protocols *string `json:"protocols"`
}

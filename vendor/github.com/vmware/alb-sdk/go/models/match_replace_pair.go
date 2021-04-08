package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MatchReplacePair match replace pair
// swagger:model MatchReplacePair
type MatchReplacePair struct {

	// String to be matched.
	// Required: true
	MatchString *string `json:"match_string"`

	// Replacement string.
	ReplacementString *ReplaceStringVar `json:"replacement_string,omitempty"`
}

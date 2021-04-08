package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HttpsecurityRule httpsecurity rule
// swagger:model HTTPSecurityRule
type HttpsecurityRule struct {

	// Action to be performed upon successful matching.
	Action *HttpsecurityAction `json:"action,omitempty"`

	// Enable or disable the rule.
	// Required: true
	Enable *bool `json:"enable"`

	// Index of the rule.
	// Required: true
	Index *int32 `json:"index"`

	// Log HTTP request upon rule match.
	Log *bool `json:"log,omitempty"`

	// Add match criteria to the rule.
	Match *MatchTarget `json:"match,omitempty"`

	// Name of the rule.
	// Required: true
	Name *string `json:"name"`
}

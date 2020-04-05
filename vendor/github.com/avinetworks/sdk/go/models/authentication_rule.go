package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthenticationRule authentication rule
// swagger:model AuthenticationRule
type AuthenticationRule struct {

	// Enable or disable authentication for matched targets. Field introduced in 18.2.5.
	Action *AuthenticationAction `json:"action,omitempty"`

	// Enable or disable the rule. Field introduced in 18.2.5.
	// Required: true
	Enable *bool `json:"enable"`

	// Index of the rule. Field introduced in 18.2.5.
	// Required: true
	Index *int32 `json:"index"`

	// Add match criteria to the rule. Field introduced in 18.2.5.
	Match *AuthenticationMatch `json:"match,omitempty"`

	// Name of the rule. Field introduced in 18.2.5.
	// Required: true
	Name *string `json:"name"`
}

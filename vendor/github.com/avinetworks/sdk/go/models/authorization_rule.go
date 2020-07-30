package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthorizationRule authorization rule
// swagger:model AuthorizationRule
type AuthorizationRule struct {

	// Authorization action when rule is matched. Field introduced in 18.2.5.
	// Required: true
	Action *AuthorizationAction `json:"action"`

	// Enable or disable the rule. Field introduced in 18.2.5.
	// Required: true
	Enable *bool `json:"enable"`

	// Index of the Authorization Policy rule. Field introduced in 18.2.5.
	// Required: true
	Index *int32 `json:"index"`

	// Authorization match criteria for the rule. Field introduced in 18.2.5.
	// Required: true
	Match *AuthorizationMatch `json:"match"`

	// Name of the rule. Field introduced in 18.2.5.
	// Required: true
	Name *string `json:"name"`
}

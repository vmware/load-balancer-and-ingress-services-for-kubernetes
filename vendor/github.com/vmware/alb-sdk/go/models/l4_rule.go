package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4Rule l4 rule
// swagger:model L4Rule
type L4Rule struct {

	// Action to be performed upon successful rule match. Field introduced in 17.2.7.
	Action *L4RuleAction `json:"action,omitempty"`

	// Enable or disable the rule. Field introduced in 17.2.7.
	Enable *bool `json:"enable,omitempty"`

	// Index of the rule. Field introduced in 17.2.7.
	// Required: true
	Index *int32 `json:"index"`

	// Match criteria of the rule. Field introduced in 17.2.7.
	Match *L4RuleMatchTarget `json:"match,omitempty"`

	// Name of the rule. Field introduced in 17.2.7.
	// Required: true
	Name *string `json:"name"`
}

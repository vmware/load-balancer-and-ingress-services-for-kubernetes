package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AttackMitigationAction attack mitigation action
// swagger:model AttackMitigationAction
type AttackMitigationAction struct {

	// Deny the attack packets further processing and drop them. Field introduced in 18.2.1.
	Deny *bool `json:"deny,omitempty"`
}

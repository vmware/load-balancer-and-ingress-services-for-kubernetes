package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeGroupOptions se group options
// swagger:model SeGroupOptions
type SeGroupOptions struct {

	// The error recovery action configured for a SE Group. Enum options - ROLLBACK_UPGRADE_OPS_ON_ERROR, SUSPEND_UPGRADE_OPS_ON_ERROR, CONTINUE_UPGRADE_OPS_ON_ERROR. Field introduced in 18.2.6.
	ActionOnError *string `json:"action_on_error,omitempty"`

	// Disable non-disruptive mechanism. Field introduced in 18.2.6.
	Disruptive *bool `json:"disruptive,omitempty"`
}

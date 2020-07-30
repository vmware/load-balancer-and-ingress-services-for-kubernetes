package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeGroupResumeOptions se group resume options
// swagger:model SeGroupResumeOptions
type SeGroupResumeOptions struct {

	// The error recovery action configured for a SE Group. Enum options - ROLLBACK_UPGRADE_OPS_ON_ERROR, SUSPEND_UPGRADE_OPS_ON_ERROR, CONTINUE_UPGRADE_OPS_ON_ERROR. Field introduced in 18.2.6.
	ActionOnError *string `json:"action_on_error,omitempty"`

	// Allow disruptive mechanism. Field introduced in 18.2.8, 20.1.1.
	Disruptive *bool `json:"disruptive,omitempty"`

	// Skip upgrade on suspended SE(s). Field introduced in 18.2.6.
	SkipSuspended *bool `json:"skip_suspended,omitempty"`
}

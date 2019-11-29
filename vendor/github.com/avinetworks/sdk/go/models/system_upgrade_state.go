package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SystemUpgradeState system upgrade state
// swagger:model SystemUpgradeState
type SystemUpgradeState struct {

	// upgrade state from controller.
	ControllerState *ControllerUpgradeState `json:"controller_state,omitempty"`

	// upgrade duration. Field introduced in 17.1.1.
	Duration *int32 `json:"duration,omitempty"`

	// upgrade end time. Field introduced in 17.1.1.
	EndTime *string `json:"end_time,omitempty"`

	// current version. Field introduced in 17.1.1.
	FromVersion *string `json:"from_version,omitempty"`

	// set if upgrade is in progress.
	// Required: true
	InProgress *bool `json:"in_progress"`

	// is set true, if patch upgrade requested by the user. Field introduced in 17.2.8.
	IsPatch *bool `json:"is_patch,omitempty"`

	// type of patch upgrade. Field introduced in 17.2.8.
	PatchType *string `json:"patch_type,omitempty"`

	// reason for upgrade failure. Field introduced in 17.1.1.
	Reason *string `json:"reason,omitempty"`

	// upgrade result. Field introduced in 17.1.1.
	Result *string `json:"result,omitempty"`

	// set if rollback is requested by the user.
	Rollback *bool `json:"rollback,omitempty"`

	// upgrade state of service engines.
	SeState *SeUpgradeStatusSummary `json:"se_state,omitempty"`

	// upgrade start time. Field introduced in 17.1.1.
	StartTime *string `json:"start_time,omitempty"`

	// version to upgrade to. Field introduced in 17.1.1.
	ToVersion *string `json:"to_version,omitempty"`
}

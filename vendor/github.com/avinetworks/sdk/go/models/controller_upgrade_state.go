package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerUpgradeState controller upgrade state
// swagger:model ControllerUpgradeState
type ControllerUpgradeState struct {

	// Number of controller_progress.
	ControllerProgress *int32 `json:"controller_progress,omitempty"`

	// Placeholder for description of property in_progress of obj type ControllerUpgradeState field type str  type boolean
	// Required: true
	InProgress *bool `json:"in_progress"`

	// notes of ControllerUpgradeState.
	Notes []string `json:"notes,omitempty"`

	// Placeholder for description of property rollback of obj type ControllerUpgradeState field type str  type boolean
	Rollback *bool `json:"rollback,omitempty"`

	//  Enum options - UPGRADE_STARTED, UPGRADE_WAITING, UPGRADE_IN_PROGRESS, UPGRADE_CONTROLLER_COMPLETED, UPGRADE_COMPLETED, UPGRADE_ABORT_IN_PROGRESS, UPGRADE_ABORTED.
	// Required: true
	State *string `json:"state"`

	// Placeholder for description of property tasks_completed of obj type ControllerUpgradeState field type str  type object
	TasksCompleted []*UpgradeTask `json:"tasks_completed,omitempty"`
}

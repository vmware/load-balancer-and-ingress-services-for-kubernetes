package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeStatusSummary se upgrade status summary
// swagger:model SeUpgradeStatusSummary
type SeUpgradeStatusSummary struct {

	// controller_version of SeUpgradeStatusSummary.
	ControllerVersion *string `json:"controller_version,omitempty"`

	// duration of SeUpgradeStatusSummary.
	Duration *string `json:"duration,omitempty"`

	// end_time of SeUpgradeStatusSummary.
	EndTime *string `json:"end_time,omitempty"`

	// Placeholder for description of property in_progress of obj type SeUpgradeStatusSummary field type str  type boolean
	InProgress *bool `json:"in_progress,omitempty"`

	// notes of SeUpgradeStatusSummary.
	Notes []string `json:"notes,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeAlreadyUpgradedAtStart []string `json:"se_already_upgraded_at_start,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeDisconnectedAtStart []string `json:"se_disconnected_at_start,omitempty"`

	// Placeholder for description of property se_group_status of obj type SeUpgradeStatusSummary field type str  type object
	SeGroupStatus []*SeGroupStatus `json:"se_group_status,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeIPMissingAtStart []string `json:"se_ip_missing_at_start,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SePoweredoffAtStart []string `json:"se_poweredoff_at_start,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeUpgradeCompleted []string `json:"se_upgrade_completed,omitempty"`

	// Placeholder for description of property se_upgrade_errors of obj type SeUpgradeStatusSummary field type str  type object
	SeUpgradeErrors []*SeUpgradeEvents `json:"se_upgrade_errors,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeUpgradeFailed []string `json:"se_upgrade_failed,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeUpgradeInProgress []string `json:"se_upgrade_in_progress,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeUpgradeNotStarted []string `json:"se_upgrade_not_started,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeUpgradeRetryCompleted []string `json:"se_upgrade_retry_completed,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeUpgradeRetryFailed []string `json:"se_upgrade_retry_failed,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeUpgradeRetryInProgress []string `json:"se_upgrade_retry_in_progress,omitempty"`

	// Service Engines that were in suspended state and were skipped upon Service Engine Group upgrade resumption. It is a reference to an object of type ServiceEngine.
	SeUpgradeSkipSuspended []string `json:"se_upgrade_skip_suspended,omitempty"`

	// Service Engines which triggered Service Engine Group to be in suspended state. It is a reference to an object of type ServiceEngine.
	SeUpgradeSuspended []string `json:"se_upgrade_suspended,omitempty"`

	// start_time of SeUpgradeStatusSummary.
	StartTime *string `json:"start_time,omitempty"`

	//  Enum options - SE_UPGRADE_PREVIEW. SE_UPGRADE_IN_PROGRESS. SE_UPGRADE_COMPLETE. SE_UPGRADE_ERROR. SE_UPGRADE_PRE_CHECKS. SE_IMAGE_INSTALL. SE_UPGRADE_IMAGE_NOT_FOUND. SE_ALREADY_UPGRADED. SE_REBOOT. SE_CONNECT_AFTER_REBOOT. SE_PRE_UPGRADE_TASKS. SE_POST_UPGRADE_TASKS. SE_WAIT_FOR_SWITCHOVER. SE_CHECK_SCALEDOUT_VS_EXISTS. SE_UPGRADE_SEMGR_REQUEST. SE_UPGRADE_SEMGR_SE_UNREACHABLE. SE_PRE_UPGRADE_SCALE_IN_OPS. SE_POST_UPGRADE_SCALE_OUT_OPS. SE_UPGRADE_SUSPENDED. SE_UPGRADE_START...
	State *string `json:"state,omitempty"`

	// Placeholder for description of property vs_errors of obj type SeUpgradeStatusSummary field type str  type object
	VsErrors []*VsError `json:"vs_errors,omitempty"`
}

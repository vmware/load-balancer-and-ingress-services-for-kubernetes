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
	SeUpgradeErrors []*SeUpgradeErrors `json:"se_upgrade_errors,omitempty"`

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

	// Service Engines that were in suspended state and were skipped upon Service Engine Group ugprade resumption. It is a reference to an object of type ServiceEngine.
	SeUpgradeSkipSuspended []string `json:"se_upgrade_skip_suspended,omitempty"`

	// Service Engines which triggered Service Engine Group to be in suspended state. It is a reference to an object of type ServiceEngine.
	SeUpgradeSuspended []string `json:"se_upgrade_suspended,omitempty"`

	// start_time of SeUpgradeStatusSummary.
	StartTime *string `json:"start_time,omitempty"`

	//  Enum options - SE_UPGRADE_PREVIEW, SE_UPGRADE_IN_PROGRESS, SE_UPGRADE_COMPLETE, SE_UPGRADE_ERROR, SE_IMAGE_INSTALL, SE_UPGRADE_IMAGE_NOT_FOUND, SE_ALREADY_UPGRADED, SE_REBOOT, SE_CONNECT_AFTER_REBOOT, SEGROUP_UPGRADE_NOT_STARTED, SEGROUP_UPGRADE_ENQUEUED, SEGROUP_UPGRADE_ENQUEUE_FAILED, SEGROUP_UPGRADE_IN_PROGRESS, SEGROUP_UPGRADE_COMPLETE, SEGROUP_UPGRADE_ERROR, SEGROUP_UPGRADE_SUSPENDED, VS_DISRUPTED, VS_SCALEIN, VS_SCALEIN_ERROR, VS_SCALEIN_ERROR_RPC_FAILED, VS_SCALEOUT, VS_SCALEOUT_ERROR, VS_SCALEOUT_ERROR_RPC_FAILED, VS_SCALEOUT_ERROR_SE_NOT_READY, VS_MIGRATE, VS_MIGRATE_ERROR, VS_MIGRATE_BACK, VS_MIGRATE_BACK_ERROR, VS_MIGRATE_BACK_NOT_NEEDED, VS_MIGRATE_ERROR_NO_CANDIDATE_SE, VS_MIGRATE_ERROR_RPC_FAILED, VS_MIGRATE_BACK_ERROR_SE_NOT_READY, VS_MIGRATE_BACK_ERROR_RPC_FAILED.
	State *string `json:"state,omitempty"`

	// Placeholder for description of property vs_errors of obj type SeUpgradeStatusSummary field type str  type object
	VsErrors []*VsError `json:"vs_errors,omitempty"`
}

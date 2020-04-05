package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeGroupStatus se group status
// swagger:model SeGroupStatus
type SeGroupStatus struct {

	// Controller version. Field introduced in 18.2.6.
	ControllerVersion *string `json:"controller_version,omitempty"`

	//  It is a reference to an object of type VirtualService.
	DisruptedVsRef []string `json:"disrupted_vs_ref,omitempty"`

	// duration of SeGroupStatus.
	Duration *string `json:"duration,omitempty"`

	// end_time of SeGroupStatus.
	EndTime *string `json:"end_time,omitempty"`

	// enqueue_time of SeGroupStatus.
	EnqueueTime *string `json:"enqueue_time,omitempty"`

	//  Enum options - HA_MODE_SHARED_PAIR, HA_MODE_SHARED, HA_MODE_LEGACY_ACTIVE_STANDBY.
	HaMode *string `json:"ha_mode,omitempty"`

	// ServiceEngineGroup upgrade in progress. Field introduced in 18.2.6.
	InProgress *bool `json:"in_progress,omitempty"`

	// notes of SeGroupStatus.
	Notes []string `json:"notes,omitempty"`

	// Number of num_se.
	NumSe *int32 `json:"num_se,omitempty"`

	// Number of num_se_with_no_vs.
	NumSeWithNoVs *int32 `json:"num_se_with_no_vs,omitempty"`

	// Number of num_se_with_vs_not_scaledout.
	NumSeWithVsNotScaledout *int32 `json:"num_se_with_vs_not_scaledout,omitempty"`

	// Number of num_se_with_vs_scaledout.
	NumSeWithVsScaledout *int32 `json:"num_se_with_vs_scaledout,omitempty"`

	// Number of num_vs.
	NumVs *int32 `json:"num_vs,omitempty"`

	// Number of num_vs_disrupted.
	NumVsDisrupted *int32 `json:"num_vs_disrupted,omitempty"`

	// Number of progress.
	Progress *int32 `json:"progress,omitempty"`

	// reason of SeGroupStatus.
	Reason []string `json:"reason,omitempty"`

	// request_time of SeGroupStatus.
	RequestTime *string `json:"request_time,omitempty"`

	// ServiceEngines are already upgraded before the upgrade. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SeAlreadyUpgradedAtStart []string `json:"se_already_upgraded_at_start,omitempty"`

	// ServiceEngines in disconnected state before starting the upgrade. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SeDisconnectedAtStart []string `json:"se_disconnected_at_start,omitempty"`

	// se_group_name of SeGroupStatus.
	SeGroupName *string `json:"se_group_name,omitempty"`

	// Unique object identifier of se_group.
	SeGroupUUID *string `json:"se_group_uuid,omitempty"`

	// ServiceEngines local ip not present before the upgrade. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SeIPMissingAtStart []string `json:"se_ip_missing_at_start,omitempty"`

	// ServiceEngines in poweredoff state before the upgrade. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SePoweredoffAtStart []string `json:"se_poweredoff_at_start,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeRebootInProgressRef *string `json:"se_reboot_in_progress_ref,omitempty"`

	// ServiceEngines upgrade completed. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SeUpgradeCompleted []string `json:"se_upgrade_completed,omitempty"`

	// ServiceEngineGroup upgrade errors. Field introduced in 18.2.6.
	SeUpgradeErrors []*SeUpgradeEvents `json:"se_upgrade_errors,omitempty"`

	// ServiceEngines upgrade failed. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SeUpgradeFailed []string `json:"se_upgrade_failed,omitempty"`

	// ServiceEngines upgrade in progress. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SeUpgradeInProgress []string `json:"se_upgrade_in_progress,omitempty"`

	// ServiceEngines upgrade not started. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.6.
	SeUpgradeNotStarted []string `json:"se_upgrade_not_started,omitempty"`

	// Service Engines that were in suspended state and were skipped upon Service Engine Group ugprade resumption. It is a reference to an object of type ServiceEngine.
	SeUpgradeSkipSuspended []string `json:"se_upgrade_skip_suspended,omitempty"`

	// Service Engines which triggered Service Engine Group to be in suspended state. It is a reference to an object of type ServiceEngine.
	SeUpgradeSuspended []string `json:"se_upgrade_suspended,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeWithNoVs []string `json:"se_with_no_vs,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeWithVsNotScaledout []string `json:"se_with_vs_not_scaledout,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeWithVsScaledout []string `json:"se_with_vs_scaledout,omitempty"`

	// start_time of SeGroupStatus.
	StartTime *string `json:"start_time,omitempty"`

	//  Enum options - SE_UPGRADE_PREVIEW. SE_UPGRADE_IN_PROGRESS. SE_UPGRADE_COMPLETE. SE_UPGRADE_ERROR. SE_UPGRADE_PRE_CHECKS. SE_IMAGE_INSTALL. SE_UPGRADE_IMAGE_NOT_FOUND. SE_ALREADY_UPGRADED. SE_REBOOT. SE_CONNECT_AFTER_REBOOT. SE_PRE_UPGRADE_TASKS. SE_POST_UPGRADE_TASKS. SE_WAIT_FOR_SWITCHOVER. SE_CHECK_SCALEDOUT_VS_EXISTS. SE_UPGRADE_SEMGR_REQUEST. SE_UPGRADE_SEMGR_SE_UNREACHABLE. SE_PRE_UPGRADE_SCALE_IN_OPS. SE_POST_UPGRADE_SCALE_OUT_OPS. SE_UPGRADE_SUSPENDED. SE_UPGRADE_START...
	State *string `json:"state,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// thread of SeGroupStatus.
	Thread *string `json:"thread,omitempty"`

	//  Enum options - TRAFFIC_DISRUPTED, TRAFFIC_NOT_DISRUPTED.
	TrafficStatus *string `json:"traffic_status,omitempty"`

	// VirtualService errors during the SeGroup upgrade. Field introduced in 18.2.6.
	VsErrors []*VsError `json:"vs_errors,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsMigrateInProgressRef []string `json:"vs_migrate_in_progress_ref,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsScaleinInProgressRef []string `json:"vs_scalein_in_progress_ref,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsScaleoutInProgressRef []string `json:"vs_scaleout_in_progress_ref,omitempty"`

	// worker of SeGroupStatus.
	Worker *string `json:"worker,omitempty"`
}

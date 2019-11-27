package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeGroupStatus se group status
// swagger:model SeGroupStatus
type SeGroupStatus struct {

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

	// se_group_name of SeGroupStatus.
	// Required: true
	SeGroupName *string `json:"se_group_name"`

	// Unique object identifier of se_group.
	// Required: true
	SeGroupUUID *string `json:"se_group_uuid"`

	//  It is a reference to an object of type ServiceEngine.
	SeRebootInProgressRef *string `json:"se_reboot_in_progress_ref,omitempty"`

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

	//  Enum options - SE_UPGRADE_PREVIEW, SE_UPGRADE_IN_PROGRESS, SE_UPGRADE_COMPLETE, SE_UPGRADE_ERROR, SE_IMAGE_INSTALL, SE_UPGRADE_IMAGE_NOT_FOUND, SE_ALREADY_UPGRADED, SE_REBOOT, SE_CONNECT_AFTER_REBOOT, SEGROUP_UPGRADE_NOT_STARTED, SEGROUP_UPGRADE_ENQUEUED, SEGROUP_UPGRADE_ENQUEUE_FAILED, SEGROUP_UPGRADE_IN_PROGRESS, SEGROUP_UPGRADE_COMPLETE, SEGROUP_UPGRADE_ERROR, SEGROUP_UPGRADE_SUSPENDED, VS_DISRUPTED, VS_SCALEIN, VS_SCALEIN_ERROR, VS_SCALEIN_ERROR_RPC_FAILED, VS_SCALEOUT, VS_SCALEOUT_ERROR, VS_SCALEOUT_ERROR_RPC_FAILED, VS_SCALEOUT_ERROR_SE_NOT_READY, VS_MIGRATE, VS_MIGRATE_ERROR, VS_MIGRATE_BACK, VS_MIGRATE_BACK_ERROR, VS_MIGRATE_BACK_NOT_NEEDED, VS_MIGRATE_ERROR_NO_CANDIDATE_SE, VS_MIGRATE_ERROR_RPC_FAILED, VS_MIGRATE_BACK_ERROR_SE_NOT_READY, VS_MIGRATE_BACK_ERROR_RPC_FAILED.
	State *string `json:"state,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// thread of SeGroupStatus.
	Thread *string `json:"thread,omitempty"`

	//  Enum options - TRAFFIC_DISRUPTED, TRAFFIC_NOT_DISRUPTED.
	TrafficStatus *string `json:"traffic_status,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsMigrateInProgressRef []string `json:"vs_migrate_in_progress_ref,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsScaleinInProgressRef []string `json:"vs_scalein_in_progress_ref,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsScaleoutInProgressRef []string `json:"vs_scaleout_in_progress_ref,omitempty"`

	// worker of SeGroupStatus.
	Worker *string `json:"worker,omitempty"`
}

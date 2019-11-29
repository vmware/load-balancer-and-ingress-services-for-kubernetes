package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeErrors se upgrade errors
// swagger:model SeUpgradeErrors
type SeUpgradeErrors struct {

	//  It is a reference to an object of type ServiceEngine.
	FromSeRef *string `json:"from_se_ref,omitempty"`

	// Number of num_se.
	NumSe *int32 `json:"num_se,omitempty"`

	// Number of num_se_group.
	NumSeGroup *int32 `json:"num_se_group,omitempty"`

	// Number of num_vs.
	NumVs *int32 `json:"num_vs,omitempty"`

	// reason of SeUpgradeErrors.
	Reason []string `json:"reason,omitempty"`

	//  Enum options - HA_MODE_SHARED_PAIR, HA_MODE_SHARED, HA_MODE_LEGACY_ACTIVE_STANDBY.
	SeGroupHaMode *string `json:"se_group_ha_mode,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`

	//  Enum options - SE_UPGRADE_PREVIEW, SE_UPGRADE_IN_PROGRESS, SE_UPGRADE_COMPLETE, SE_UPGRADE_ERROR, SE_IMAGE_INSTALL, SE_UPGRADE_IMAGE_NOT_FOUND, SE_ALREADY_UPGRADED, SE_REBOOT, SE_CONNECT_AFTER_REBOOT, SEGROUP_UPGRADE_NOT_STARTED, SEGROUP_UPGRADE_ENQUEUED, SEGROUP_UPGRADE_ENQUEUE_FAILED, SEGROUP_UPGRADE_IN_PROGRESS, SEGROUP_UPGRADE_COMPLETE, SEGROUP_UPGRADE_ERROR, SEGROUP_UPGRADE_SUSPENDED, VS_DISRUPTED, VS_SCALEIN, VS_SCALEIN_ERROR, VS_SCALEIN_ERROR_RPC_FAILED, VS_SCALEOUT, VS_SCALEOUT_ERROR, VS_SCALEOUT_ERROR_RPC_FAILED, VS_SCALEOUT_ERROR_SE_NOT_READY, VS_MIGRATE, VS_MIGRATE_ERROR, VS_MIGRATE_BACK, VS_MIGRATE_BACK_ERROR, VS_MIGRATE_BACK_NOT_NEEDED, VS_MIGRATE_ERROR_NO_CANDIDATE_SE, VS_MIGRATE_ERROR_RPC_FAILED, VS_MIGRATE_BACK_ERROR_SE_NOT_READY, VS_MIGRATE_BACK_ERROR_RPC_FAILED.
	// Required: true
	Task *string `json:"task"`

	//  It is a reference to an object of type ServiceEngine.
	ToSeRef *string `json:"to_se_ref,omitempty"`

	//  Enum options - TRAFFIC_DISRUPTED, TRAFFIC_NOT_DISRUPTED.
	TrafficStatus *string `json:"traffic_status,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsRef *string `json:"vs_ref,omitempty"`
}

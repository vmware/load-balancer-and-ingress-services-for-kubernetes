package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeEvents se upgrade events
// swagger:model SeUpgradeEvents
type SeUpgradeEvents struct {

	//  It is a reference to an object of type ServiceEngine.
	FromSeRef *string `json:"from_se_ref,omitempty"`

	// Number of num_se.
	NumSe *int32 `json:"num_se,omitempty"`

	// Number of num_se_group.
	NumSeGroup *int32 `json:"num_se_group,omitempty"`

	// Number of num_vs.
	NumVs *int32 `json:"num_vs,omitempty"`

	// reason of SeUpgradeEvents.
	Reason []string `json:"reason,omitempty"`

	//  Enum options - HA_MODE_SHARED_PAIR, HA_MODE_SHARED, HA_MODE_LEGACY_ACTIVE_STANDBY.
	SeGroupHaMode *string `json:"se_group_ha_mode,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`

	// List of sub_tasks executed. Field introduced in 20.1.4.
	SubTasks []string `json:"sub_tasks,omitempty"`

	//  Enum options - SE_UPGRADE_PREVIEW. SE_UPGRADE_IN_PROGRESS. SE_UPGRADE_COMPLETE. SE_UPGRADE_ERROR. SE_UPGRADE_PRE_CHECKS. SE_IMAGE_INSTALL. SE_UPGRADE_IMAGE_NOT_FOUND. SE_ALREADY_UPGRADED. SE_REBOOT. SE_CONNECT_AFTER_REBOOT. SE_PRE_UPGRADE_TASKS. SE_POST_UPGRADE_TASKS. SE_WAIT_FOR_SWITCHOVER. SE_CHECK_SCALEDOUT_VS_EXISTS. SE_UPGRADE_SEMGR_REQUEST. SE_UPGRADE_SEMGR_SE_UNREACHABLE. SE_PRE_UPGRADE_SCALE_IN_OPS. SE_POST_UPGRADE_SCALE_OUT_OPS. SE_UPGRADE_SUSPENDED. SE_UPGRADE_START...
	Task *string `json:"task,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	ToSeRef *string `json:"to_se_ref,omitempty"`

	//  Enum options - TRAFFIC_DISRUPTED, TRAFFIC_NOT_DISRUPTED.
	TrafficStatus *string `json:"traffic_status,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsRef *string `json:"vs_ref,omitempty"`
}

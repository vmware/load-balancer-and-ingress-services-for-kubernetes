package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ScaleStatus scale status
// swagger:model ScaleStatus
type ScaleStatus struct {

	//  Enum options - OTHER, CREATE, READ, UPDATE, DELETE, SCALE_OUT, SCALE_IN, SE_REMOVED, SE_DISCONNECT, SE_RECONNECT, WARM_RESTART, COLD_RESTART, UPDATE_LOGMGR_MAP, MIGRATE_SCALEOUT, MIGRATE_SCALEIN, INITIAL_PLACEMENT, ROTATE_KEYS, GLB_MGR_UPDATE, UPDATE_DNS_RECORDS, SCALEOUT_ADMINUP, SCALEIN_ADMINDOWN, SCALEOUT_READY, SCALEIN_READY, SE_PRERELEASED, SE_ADDED, SE_PRERELEASE, SE_SCALING_IN, SE_DELETING, SE_PRERELEASE_FAILED, SE_SCALEOUT_FAILED, SCALEOUT_ROLLEDBACK, RSRC_UPDATE, SE_UPDATED, GLB_MGR_UPDATE_GS_STATUS, NEW_PRIMARY, SE_FORCE_RELEASE, SWITCHOVER, SE_ASSIGN_NO_CHANGE, GLB_MGR_DNS_GEO_UPDATE, GLB_MGR_DNS_CLEANUP, NEW_STANDBY, CONNECTED, NOT_CONNECTED, NOT_AT_CURR_VERSION, AT_CURR_VERSION, ATTACH_IP_SUCCESS, SEGROUP_CHANGED, VS_ENABLED, VS_DISABLED, UPDATE_DURING_WARMSTART, RSRC_UPDATE_DURING_WARMSTART, WARMSTART_PARENT_NOT_FOUND, WARMSTART_PARENT_SELIST_MISMATCH, WARMSTART_PARENT_SEGROUP_MISMATCH, WARMSTART_RESYNC_SENT, WARMSTART_RESYNC_RESPONSE, RPC_TO_RESMGR_FAILED, SCALEOUT_READY_IGNORED, SCALEOUT_READY_TIMEDOUT. Field introduced in 17.1.1.
	Action *string `json:"action,omitempty"`

	//  Field introduced in 18.1.1.
	ActionSuccess *bool `json:"action_success,omitempty"`

	// end_time_str of ScaleStatus.
	EndTimeStr *string `json:"end_time_str,omitempty"`

	// Number of num_se_assigned.
	NumSeAssigned *int32 `json:"num_se_assigned,omitempty"`

	// Number of num_se_requested.
	NumSeRequested *int32 `json:"num_se_requested,omitempty"`

	// reason of ScaleStatus.
	Reason []string `json:"reason,omitempty"`

	// Number of reason_code.
	ReasonCode *int64 `json:"reason_code,omitempty"`

	// reason_code_string of ScaleStatus.
	ReasonCodeString *string `json:"reason_code_string,omitempty"`

	// scale_se of ScaleStatus.
	ScaleSe *string `json:"scale_se,omitempty"`

	// start_time_str of ScaleStatus.
	StartTimeStr *string `json:"start_time_str,omitempty"`

	//  Enum options - SCALEOUT_PROCESSING, SCALEOUT_AWAITING_SE_ASSIGNMENT, SCALEOUT_CREATING_SE, SCALEOUT_RESOURCES, SCALEOUT_AWAITING_SE_PROGRAMMING, SCALEOUT_WAIT_FOR_SE_READY, SCALEOUT_SUCCESS, SCALEOUT_ERROR, SCALEOUT_ROLLBACK, SCALEOUT_ERROR_DISABLED, SCALEIN_AWAITING_SE_PRE_RELEASE, SCALEIN_AWAITING_SE_PROGRAMMING, SCALEIN_WAIT_FOR_SE_READY, SCALEIN_AWAITING_SE_RELEASE, SCALEIN_SUCCESS, SCALEIN_ERROR, MIGRATE_SCALEOUT_AWAITING_SE_ASSIGNMENT, MIGRATE_SCALEOUT_CREATING_SE, MIGRATE_SCALEOUT_RESOURCES, MIGRATE_SCALEOUT_AWAITING_SE_PROGRAMMING, MIGRATE_SCALEOUT_WAIT_FOR_SE_READY, MIGRATE_SCALEOUT_SUCCESS, MIGRATE_SCALEOUT_ROLLBACK, MIGRATE_SCALEOUT_ERROR, MIGRATE_SCALEIN_AWAITING_SE_PRE_RELEASE, MIGRATE_SCALEIN_AWAITING_SE_PROGRAMMING, MIGRATE_SCALEIN_WAIT_FOR_SE_READY, MIGRATE_SCALEIN_AWAITING_SE_RELEASE, MIGRATE_SCALEIN_SUCCESS, MIGRATE_SCALEIN_ERROR, MIGRATE_SUCCESS, MIGRATE_ERROR.
	State *string `json:"state,omitempty"`

	// Placeholder for description of property vip_placement_resolution_info of obj type ScaleStatus field type str  type object
	VipPlacementResolutionInfo *VipPlacementResolutionInfo `json:"vip_placement_resolution_info,omitempty"`
}

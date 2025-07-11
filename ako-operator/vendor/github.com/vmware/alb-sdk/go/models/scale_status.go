// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ScaleStatus scale status
// swagger:model ScaleStatus
type ScaleStatus struct {

	//  Enum options - OTHER, CREATE, READ, UPDATE, DELETE, SCALE_OUT, SCALE_IN, SE_REMOVED, SE_DISCONNECT, SE_RECONNECT, WARM_RESTART, COLD_RESTART, UPDATE_LOGMGR_MAP, MIGRATE_SCALEOUT, MIGRATE_SCALEIN, INITIAL_PLACEMENT, ROTATE_KEYS, GLB_MGR_UPDATE, UPDATE_DNS_RECORDS, SCALEOUT_ADMINUP, SCALEIN_ADMINDOWN, SCALEOUT_READY, SCALEIN_READY, SE_PRERELEASED, SE_ADDED, SE_PRERELEASE, SE_SCALING_IN, SE_DELETING, SE_PRERELEASE_FAILED, SE_SCALEOUT_FAILED, SCALEOUT_ROLLEDBACK, RSRC_UPDATE, SE_UPDATED, GLB_MGR_UPDATE_GS_STATUS, NEW_PRIMARY, SE_FORCE_RELEASE, SWITCHOVER, SE_ASSIGN_NO_CHANGE, GLB_MGR_DNS_GEO_UPDATE, GLB_MGR_DNS_CLEANUP, NEW_STANDBY, CONNECTED, NOT_CONNECTED, NOT_AT_CURR_VERSION, AT_CURR_VERSION, ATTACH_IP_SUCCESS, SEGROUP_CHANGED, VS_ENABLED, VS_DISABLED, UPDATE_DURING_WARMSTART, RSRC_UPDATE_DURING_WARMSTART, WARMSTART_PARENT_NOT_FOUND, WARMSTART_PARENT_SELIST_MISMATCH, WARMSTART_PARENT_SEGROUP_MISMATCH, WARMSTART_RESYNC_SENT, WARMSTART_RESYNC_RESPONSE, RPC_TO_RESMGR_FAILED, SCALEOUT_READY_IGNORED, SCALEOUT_READY_TIMEDOUT, SCALEOUT_READY_DISCONNECTED, FORCE_SCALEIN_POST_WARMSTART, MIGRATE_SCALEIN_SKIPPED, MIGRATE_DISRUPTED, APIC_PLACEMENT, SE_VIP_MAC_UPDATE, MIN_SCALEOUT_UPDATED, VIP_CHANGE_DISRUPTIVE, POOL_CHANGE_DISRUPTIVE, ECMP_CHANGE_DISRUPTIVE, ENABLE_RHI_CHANGE_DISRUPTIVE, VIP_AS_SNAT_CHANGE_DISRUPTIVE, SE_GROUP_CHANGE_DISRUPTIVE, PARENT_VS_CHANGE_DISRUPTIVE, SNAT_POOL_CHANGE_DISRUPTIVE, UNSET_USE_VIP_AS_SNAT, SE_MGMTIP_CHANGE, TRAFFIC_ENABLED, TRAFFIC_DISABLED, SCALEIN_READY_TIMEDOUT, SE_READY, SE_READY_TIMEDOUT, NEW_PRIMARY_READY, NEW_PRIMARY_READY_TIMEDOUT, ATTACH_IP_IN_PROG, ATTACH_IP_TIMEDOUT, DETACH_IP_SUCCESS, DETACH_IP_IN_PROG, DETACH_IP_TIMEDOUT, CLEAR_ADMIN_DOWN, MULTIPLE_PRIMARY, BGP_PEER_LABELS_CHANGE_DISRUPTIVE. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Action *string `json:"action,omitempty"`

	//  Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActionSuccess *bool `json:"action_success,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EndTimeStr *string `json:"end_time_str,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSeAssigned *uint32 `json:"num_se_assigned,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSeRequested *uint32 `json:"num_se_requested,omitempty"`

	//  Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	PlacementReadFailCnt *uint32 `json:"placement_read_fail_cnt,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason []string `json:"reason,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReasonCode *uint64 `json:"reason_code,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReasonCodeString *string `json:"reason_code_string,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleSe *string `json:"scale_se,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StartTimeStr *string `json:"start_time_str,omitempty"`

	//  Enum options - SCALEOUT_PROCESSING, SCALEOUT_AWAITING_SE_ASSIGNMENT, SCALEOUT_AWAITING_ADMINUP, SCALEOUT_CREATING_SE, SCALEOUT_ADMINUP_AWAITING_CLOUD_ATTACH, SCALEOUT_RESOURCES, SCALEOUT_AWAITING_CLOUD_ATTACH, SCALEOUT_AWAITING_SE_PROGRAMMING, SCALEOUT_AWAITING_SE_READY, SCALEOUT_WAIT_FOR_SE_READY, SCALEOUT_SUCCESS, SCALEOUT_ERROR, SCALEOUT_ROLLBACK, SCALEOUT_ERROR_DISABLED, SCALEIN_AWAITING_SE_READY, SCALEIN_AWAITING_SE_PRE_RELEASE, SCALEIN_AWAITING_PRIMARY_SWITCHOVER, SCALEIN_AWAITING_SE_PROGRAMMING, SCALEIN_AWAITING_CLOUD_DETACH, SCALEIN_WAIT_FOR_SE_READY, SCALEIN_AWAITING_SE_RELEASE, SCALEIN_SUCCESS, SCALEIN_AWAITING_PRIMARY_ATTACH, SCALEIN_ERROR, SCALEIN_ADMINDOWN_AWAITING_CLOUD_DETACH, MIGRATE_SCALEOUT_AWAITING_SE_ASSIGNMENT, MIGRATE_SCALEOUT_CREATING_SE, MIGRATE_SCALEOUT_AWAITING_CLOUD_ATTACH, MIGRATE_SCALEOUT_RESOURCES, MIGRATE_SCALEOUT_AWAITING_SE_PROGRAMMING, MIGRATE_SCALEOUT_WAIT_FOR_SE_READY, MIGRATE_SCALEOUT_AWAITING_SE_READY, MIGRATE_SCALEOUT_SUCCESS, MIGRATE_SCALEOUT_ROLLBACK, MIGRATE_SCALEOUT_ERROR, MIGRATE_SCALEIN_AWAITING_SE_PRE_RELEASE, MIGRATE_SCALEIN_AWAITING_SE_PROGRAMMING, MIGRATE_SCALEIN_WAIT_FOR_SE_READY, MIGRATE_SCALEIN_AWAITING_CLOUD_DETACH, MIGRATE_SCALEIN_AWAITING_SE_RELEASE, MIGRATE_SCALEIN_AWAITING_PRIMARY_SWITCHOVER, MIGRATE_SCALEIN_SUCCESS, MIGRATE_SCALEIN_ERROR, MIGRATE_SCALEIN_AWAITING_PRIMARY_ATTACH, MIGRATE_SUCCESS, MIGRATE_ERROR, MIGRATE_SCALEIN_AWAITING_SE_READY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipPlacementResolutionInfo *VipPlacementResolutionInfo `json:"vip_placement_resolution_info,omitempty"`
}

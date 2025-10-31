// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeUpgradeEvents se upgrade events
// swagger:model SeUpgradeEvents
type SeUpgradeEvents struct {

	//  It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FromSeRef *string `json:"from_se_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSe *int32 `json:"num_se,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSeGroup *int32 `json:"num_se_group,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumVs *int32 `json:"num_vs,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason []string `json:"reason,omitempty"`

	//  Enum options - HA_MODE_SHARED_PAIR, HA_MODE_SHARED, HA_MODE_LEGACY_ACTIVE_STANDBY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupHaMode *string `json:"se_group_ha_mode,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	//  It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`

	// List of sub_tasks executed. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SubTasks []string `json:"sub_tasks,omitempty"`

	//  Enum options - SE_UPGRADE_PREVIEW, SE_UPGRADE_IN_PROGRESS, SE_UPGRADE_COMPLETE, SE_UPGRADE_ERROR, SE_UPGRADE_PRE_CHECKS, SE_IMAGE_INSTALL, SE_UPGRADE_IMAGE_NOT_FOUND, SE_ALREADY_UPGRADED, SE_REBOOT, SE_CONNECT_AFTER_REBOOT, SE_PRE_UPGRADE_TASKS, SE_POST_UPGRADE_TASKS, SE_WAIT_FOR_SWITCHOVER, SE_CHECK_SCALEDOUT_VS_EXISTS, SE_UPGRADE_SEMGR_REQUEST, SE_UPGRADE_SEMGR_SE_UNREACHABLE, SE_PRE_UPGRADE_SCALE_IN_OPS, SE_POST_UPGRADE_SCALE_OUT_OPS, SE_UPGRADE_SUSPENDED, SE_UPGRADE_START, SE_UPGRADE_PAUSED, SE_UPGRADE_FAILED, SE_UPGRADE_VERSION_CHECKS, SE_UPGRADE_CONNECTIVITY_CHECKS, SE_UPGRADE_VERIFY_VERSION, SE_UPGRADE_SKIP_RESUME_OPS, SE_UPGRADE_SEMGR_DONE, SEGROUP_UPGRADE_NOT_STARTED, SEGROUP_UPGRADE_ENQUEUED, SEGROUP_UPGRADE_ENQUEUE_FAILED, SEGROUP_UPGRADE_IN_PROGRESS, SEGROUP_UPGRADE_COMPLETE, SEGROUP_UPGRADE_ERROR, SEGROUP_UPGRADE_SUSPENDED, VS_DISRUPTED, VS_SCALEIN, VS_SCALEIN_ERROR, VS_SCALEIN_ERROR_RPC_FAILED, VS_SCALEOUT, VS_SCALEOUT_ERROR, VS_SCALEOUT_ERROR_RPC_FAILED, VS_SCALEOUT_ERROR_SE_NOT_READY, VS_MIGRATE, VS_MIGRATE_ERROR, VS_MIGRATE_BACK, VS_MIGRATE_BACK_ERROR, VS_MIGRATE_BACK_NOT_NEEDED, VS_MIGRATE_ERROR_NO_CANDIDATE_SE, VS_MIGRATE_ERROR_RPC_FAILED, VS_MIGRATE_BACK_ERROR_SE_NOT_READY, VS_MIGRATE_BACK_ERROR_RPC_FAILED, SEGROUP_PAUSE_PLACEMENT, SEGROUP_RESUME_PLACEMENT, SEGROUP_CLOUD_DISCOVERY, SEGROUP_IMAGE_GENERATION, SEGROUP_IMAGE_COPY_INSTALL_TO_SES, SEGROUP_SERIAL_SE_UPGRADE, SEGROUP_PARALLEL_SE_UPGRADE, SEGROUP_V2_TO_V1_ROLLBACK, SEGROUP_FAILED_SE_ERROR_RECOVERY, SEGROUP_SE_CONNECTIVITY_CHECKS, SEGROUP_UPGRADE_START, SEGROUP_WAIT_FOR_WARM_START_DONE, SEGROUP_PRE_SNAPSHOT, SEGROUP_POST_SNAPSHOT, SEGROUP_WAIT_FOR_SNAPSHOT_COLLECTION. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Task *string `json:"task,omitempty"`

	//  It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ToSeRef *string `json:"to_se_ref,omitempty"`

	//  Enum options - TRAFFIC_DISRUPTED, TRAFFIC_NOT_DISRUPTED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TrafficStatus *string `json:"traffic_status,omitempty"`

	//  It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsRef *string `json:"vs_ref,omitempty"`
}

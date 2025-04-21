// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UpgradeStatusInfo upgrade status info
// swagger:model UpgradeStatusInfo
type UpgradeStatusInfo struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Backward compatible abort function name. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AfterRebootRollbackFnc *string `json:"after_reboot_rollback_fnc,omitempty"`

	// Backward compatible task dict name. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AfterRebootTaskName *string `json:"after_reboot_task_name,omitempty"`

	// Flag for clean installation. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Clean *bool `json:"clean,omitempty"`

	// Duration of Upgrade operation in seconds. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Duration *int32 `json:"duration,omitempty"`

	// Check if the patch rollback is possible on this node. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnablePatchRollback *bool `json:"enable_patch_rollback,omitempty"`

	// Check if the rollback is possible on this node. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableRollback *bool `json:"enable_rollback,omitempty"`

	// End time of Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EndTime *string `json:"end_time,omitempty"`

	// Enqueue time of Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnqueueTime *string `json:"enqueue_time,omitempty"`

	// Fips mode for the entire system. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FipsMode *bool `json:"fips_mode,omitempty"`

	// Record of past operations on this node. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	History []*OpsHistory `json:"history,omitempty"`

	// Image path of current base image. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ImagePath *string `json:"image_path,omitempty"`

	// Image uuid for identifying the current base image. It is a reference to an object of type Image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ImageRef *string `json:"image_ref,omitempty"`

	// Name of the system such as cluster name, se group name and se name. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Type of the system such as controller_cluster, se_group or se. Enum options - NODE_CONTROLLER_CLUSTER, NODE_SE_GROUP, NODE_SE_TYPE. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeType *string `json:"node_type,omitempty"`

	// Cloud that this object belongs to. It is a reference to an object of type Cloud. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjCloudRef *string `json:"obj_cloud_ref,omitempty"`

	// Parameters associated with the Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Params *UpgradeOpsParam `json:"params,omitempty"`

	// Image path of current patch image. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchImagePath *string `json:"patch_image_path,omitempty"`

	// Image uuid for identifying the current patch.Example  Base-image is 18.2.6 and a patch 6p1 is applied, then this field will indicate the 6p1 value. . It is a reference to an object of type Image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchImageRef *string `json:"patch_image_ref,omitempty"`

	// List of patches applied to this node. Example  Base-image is 18.2.6 and a patch 6p1 is applied, then a patch 6p5 applied. This field will indicate the [{'6p1', '6p1_image_uuid'}, {'6p5', '6p5_image_uuid'}] value. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchList []*PatchData `json:"patch_list,omitempty"`

	// Flag for patch op with reboot. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchReboot *bool `json:"patch_reboot,omitempty"`

	// Current patch version applied to this node. Example  Base-image is 18.2.6 and a patch 6p1 is applied, then this field will indicate the 6p1 value. . Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchVersion *string `json:"patch_version,omitempty"`

	// Image path of previous base image. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrevImagePath *string `json:"prev_image_path,omitempty"`

	// Image path of previous patch image. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrevPatchImagePath *string `json:"prev_patch_image_path,omitempty"`

	// Remote image reference of previous base image. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PrevRemoteImageRef *string `json:"prev_remote_image_ref,omitempty"`

	// Image uuid for identifying previous base image.Example  Base-image was 18.2.5 and an upgrade was done to 18.2.6, then this field will indicate the 18.2.5 value. . It is a reference to an object of type Image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviousImageRef *string `json:"previous_image_ref,omitempty"`

	// Image uuid for identifying previous patch.Example  Base-image was 18.2.6 with a patch 6p1. Upgrade was initiated to 18.2.8 with patch 8p1. The previous_image field will contain 18.2.6 and this field will indicate the 6p1 value. . It is a reference to an object of type Image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviousPatchImageRef *string `json:"previous_patch_image_ref,omitempty"`

	// List of patches applied to this node on previous major version. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviousPatchList []*PatchData `json:"previous_patch_list,omitempty"`

	// Previous patch version applied to this node.Example  Base-image was 18.2.6 with a patch 6p1. Upgrade was initiated to 18.2.8 with patch 8p1. The previous_image field will contain 18.2.6 and this field will indicate the 6p1 value. . Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviousPatchVersion *string `json:"previous_patch_version,omitempty"`

	// Previous version prior to upgrade.Example  Base-image was 18.2.5 and an upgrade was done to 18.2.6, then this field will indicate the 18.2.5 value. . Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviousVersion *string `json:"previous_version,omitempty"`

	// Upgrade operations progress which holds value between 0-100. Allowed values are 0-100. Field introduced in 18.2.8, 20.1.1. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Progress uint32 `json:"progress,omitempty"`

	// Descriptive reason for the Upgrade state. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Remote image reference of current base image. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RemoteImageRef *string `json:"remote_image_ref,omitempty"`

	// Image path of se patch image.(required in case of reimage and upgrade + patch). Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePatchImagePath *string `json:"se_patch_image_path,omitempty"`

	// Image uuid for identifying the current se patch required in case of system upgrade(re-image) with se patch. . It is a reference to an object of type Image. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePatchImageRef *string `json:"se_patch_image_ref,omitempty"`

	// ServiceEngineGroup upgrade errors. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeUpgradeEvents []*SeUpgradeEvents `json:"se_upgrade_events,omitempty"`

	// se_patch may be different from the controller_patch. It has to be saved in the journal for subsequent consumption. The SeGroup params will be saved in the controller entry as seg_params. . Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SegParams *UpgradeOpsParam `json:"seg_params,omitempty"`

	// Detailed SeGroup status. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SegStatus *SeGroupStatus `json:"seg_status,omitempty"`

	// Start time of Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StartTime *string `json:"start_time,omitempty"`

	// Current status of the Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *UpgradeOpsState `json:"state,omitempty"`

	// Record of Pre/Post snapshot captured for current upgrade operation. It is a reference to an object of type StatediffOperation. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StatediffRef *string `json:"statediff_ref,omitempty"`

	// Flag is set only in the cluster if the upgrade is initiated as a system-upgrade. . Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	System *bool `json:"system,omitempty"`

	// Completed set of tasks in the Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TasksCompleted *int32 `json:"tasks_completed,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Total number of tasks in the Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TotalTasks *int32 `json:"total_tasks,omitempty"`

	// Events performed for Upgrade operation. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpgradeEvents []*EventMap `json:"upgrade_events,omitempty"`

	// Upgrade operations requested. Enum options - UPGRADE, PATCH, ROLLBACK, ROLLBACKPATCH, SEGROUP_RESUME, EVAL_UPGRADE, EVAL_PATCH, EVAL_ROLLBACK, EVAL_ROLLBACKPATCH, EVAL_SEGROUP_RESUME. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpgradeOps *string `json:"upgrade_ops,omitempty"`

	// Upgrade readiness check execution detail. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpgradeReadiness *UpgradeReadinessCheckObj `json:"upgrade_readiness,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID Identifier for the system such as cluster, se group and se. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Current base image applied to this node. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *string `json:"version,omitempty"`
}

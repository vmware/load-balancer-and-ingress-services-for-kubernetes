// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UpgradeReadinessCheckObj upgrade readiness check obj
// swagger:model UpgradeReadinessCheckObj
type UpgradeReadinessCheckObj struct {

	// List of Upgrade readiness check exceptions. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Checks []*MustChecksInfo `json:"checks,omitempty"`

	// No. of checks completed. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ChecksCompleted *int32 `json:"checks_completed,omitempty"`

	// Time taken to complete upgrade readiness checks in seconds. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Duration uint32 `json:"duration,omitempty"`

	// Time at which execution of upgrade readiness checks was completed. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndTime *string `json:"end_time,omitempty"`

	// Image uuid for identifying the next base image. It is a reference to an object of type Image. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImageRef *string `json:"image_ref,omitempty"`

	// Image uuid for identifying the next patch. It is a reference to an object of type Image. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PatchImageRef *string `json:"patch_image_ref,omitempty"`

	// Time at which execution of upgrade readiness checks was started. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartTime *string `json:"start_time,omitempty"`

	// The Upgrade readiness check operations current fsm-state. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	State *UpgradeOpsState `json:"state,omitempty"`

	// Total no. of checks. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TotalChecks *int32 `json:"total_checks,omitempty"`

	// Upgrade operations along with type requested such as UpgradeSystem UpgradeController etc. Enum options - UPGRADE, PATCH, ROLLBACK, ROLLBACKPATCH, SEGROUP_RESUME, EVAL_UPGRADE, EVAL_PATCH, EVAL_ROLLBACK, EVAL_ROLLBACKPATCH, EVAL_SEGROUP_RESUME. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpgradeOps *string `json:"upgrade_ops,omitempty"`
}

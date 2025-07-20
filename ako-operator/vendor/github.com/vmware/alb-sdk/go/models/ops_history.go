// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpsHistory ops history
// swagger:model OpsHistory
type OpsHistory struct {

	// Duration of Upgrade operation in seconds. Field introduced in 20.1.4. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Duration *int32 `json:"duration,omitempty"`

	// End time of Upgrade operation. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndTime *string `json:"end_time,omitempty"`

	// Upgrade operation performed. Enum options - UPGRADE, PATCH, ROLLBACK, ROLLBACKPATCH, SEGROUP_RESUME, EVAL_UPGRADE, EVAL_PATCH, EVAL_ROLLBACK, EVAL_ROLLBACKPATCH, EVAL_SEGROUP_RESUME, EVAL_RESTORE, RESTORE. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ops *string `json:"ops,omitempty"`

	// Patch after the upgrade operation. . Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PatchVersion *string `json:"patch_version,omitempty"`

	// ServiceEngineGroup/SE events for upgrade operation. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeUpgradeEvents []*SeUpgradeEvents `json:"se_upgrade_events,omitempty"`

	// SeGroup status for the upgrade operation. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SegStatus *SeGroupStatus `json:"seg_status,omitempty"`

	// Start time of Upgrade operation. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartTime *string `json:"start_time,omitempty"`

	// Upgrade operation status. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	State *UpgradeOpsState `json:"state,omitempty"`

	// Record of Pre/Post snapshot captured for current upgrade operation. It is a reference to an object of type StatediffOperation. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StatediffRef *string `json:"statediff_ref,omitempty"`

	// Controller events for Upgrade operation. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpgradeEvents []*EventMap `json:"upgrade_events,omitempty"`

	// Image after the upgrade operation. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Version *string `json:"version,omitempty"`
}

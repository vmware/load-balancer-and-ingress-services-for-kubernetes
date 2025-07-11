// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ObjSyncConfig obj sync config
// swagger:model ObjSyncConfig
type ObjSyncConfig struct {

	// SE CPU limit for InterSE Object Distribution. Allowed values are 15-80. Field introduced in 20.1.3. Unit is PERCENT. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjsyncCPULimit *uint32 `json:"objsync_cpu_limit,omitempty"`

	// Hub election interval for InterSE Object Distribution. Allowed values are 30-300. Field introduced in 20.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjsyncHubElectInterval *uint32 `json:"objsync_hub_elect_interval,omitempty"`

	// Reconcile interval for InterSE Object Distribution. Allowed values are 1-120. Field introduced in 20.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjsyncReconcileInterval *uint32 `json:"objsync_reconcile_interval,omitempty"`
}

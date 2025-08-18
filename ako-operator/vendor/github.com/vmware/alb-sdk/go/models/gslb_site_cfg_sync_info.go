// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteCfgSyncInfo gslb site cfg sync info
// swagger:model GslbSiteCfgSyncInfo
type GslbSiteCfgSyncInfo struct {

	// Objects that could NOT be synced to the site . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErroredObjects []*VersionInfo `json:"errored_objects,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Last object having replication issue. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LastFailObj *ConfigVersionStatus `json:"last_fail_obj,omitempty"`

	// Reason for the replication issues. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Recommended way to resolve replication issue. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Recommendation *string `json:"recommendation,omitempty"`

	// Configuration sync-state of the site . Enum options - GSLB_SITE_CFG_IN_SYNC, GSLB_SITE_CFG_OUT_OF_SYNC, GSLB_SITE_CFG_SYNC_DISABLED, GSLB_SITE_CFG_SYNC_IN_PROGRESS, GSLB_SITE_CFG_SYNC_NOT_APPLICABLE, GSLB_SITE_CFG_SYNCED_TILL_CHECKPOINT, GSLB_SITE_CFG_SYNC_SUSPENDED, GSLB_SITE_CFG_SYNC_STALLED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SyncState *string `json:"sync_state,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugVirtualServiceObjSync debug virtual service obj sync
// swagger:model DebugVirtualServiceObjSync
type DebugVirtualServiceObjSync struct {

	// Triggers Initial Sync on all the SEs of this VS. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TriggerInitialSync *bool `json:"trigger_initial_sync,omitempty"`
}

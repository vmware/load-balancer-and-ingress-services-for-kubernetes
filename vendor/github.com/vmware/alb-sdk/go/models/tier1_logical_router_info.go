// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Tier1LogicalRouterInfo tier1 logical router info
// swagger:model Tier1LogicalRouterInfo
type Tier1LogicalRouterInfo struct {

	// Overlay segment path. Example- /infra/segments/Seg-Web-T1-01. Field introduced in 20.1.1.
	SegmentID *string `json:"segment_id,omitempty"`

	// Tier1 logical router path. Example- /infra/tier-1s/T1-01. Field introduced in 20.1.1.
	// Required: true
	Tier1LrID *string `json:"tier1_lr_id"`
}

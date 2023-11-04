// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtTier1SegmentManualMode nsxt tier1 segment manual mode
// swagger:model NsxtTier1SegmentManualMode
type NsxtTier1SegmentManualMode struct {

	// Tier1 logical router placement information. Field introduced in 20.1.1. Minimum of 1 items required. Maximum of 300 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tier1Lrs []*Tier1LogicalRouterInfo `json:"tier1_lrs,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtTier1SegmentAutomaticMode nsxt tier1 segment automatic mode
// swagger:model NsxtTier1SegmentAutomaticMode
type NsxtTier1SegmentAutomaticMode struct {

	// Uber IP subnet for the logical segments created automatically by Avi controller. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NsxtSegmentSubnet *IPAddrPrefix `json:"nsxt_segment_subnet"`

	// The number of SE data vNic's that can be connected to the Avi logical segment. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSePerSegment *uint32 `json:"num_se_per_segment,omitempty"`

	// Tier1 logical router IDs. Field introduced in 20.1.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tier1LrIds []string `json:"tier1_lr_ids,omitempty"`
}

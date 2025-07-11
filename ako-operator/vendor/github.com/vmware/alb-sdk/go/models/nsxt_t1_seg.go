// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtT1Seg nsxt t1 seg
// swagger:model NsxtT1Seg
type NsxtT1Seg struct {

	// NSX-T Segment. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Segment *string `json:"segment,omitempty"`

	// NSX-T Tier1. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Tier1 *string `json:"tier1,omitempty"`
}

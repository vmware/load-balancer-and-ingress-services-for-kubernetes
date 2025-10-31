// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtT1SegDetails nsxt t1 seg details
// swagger:model NsxtT1SegDetails
type NsxtT1SegDetails struct {

	// NSX-T cloud id. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CcID *string `json:"cc_id,omitempty"`

	// Error message. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorString *string `json:"error_string,omitempty"`

	// NSX-T tier1(s) segment(s). Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	T1seg []*NsxtT1Seg `json:"t1seg,omitempty"`
}

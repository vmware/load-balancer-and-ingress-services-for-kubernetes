// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsRuntimeSummary vs runtime summary
// swagger:model VsRuntimeSummary
type VsRuntimeSummary struct {

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PercentSesUp *int32 `json:"percent_ses_up,omitempty"`

	// Vip summary of the virtual service. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VipSummary *VipSummary `json:"vip_summary,omitempty"`
}

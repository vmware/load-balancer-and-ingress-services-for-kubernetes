// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsvipRuntimeSummary vsvip runtime summary
// swagger:model VsvipRuntimeSummary
type VsvipRuntimeSummary struct {

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumSeAssigned *uint32 `json:"num_se_assigned,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumSeRequested *uint32 `json:"num_se_requested,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PercentSesUp *int32 `json:"percent_ses_up,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceEngine []*VipSeAssigned `json:"service_engine,omitempty"`

	// This field is used to uniquely identify the vip. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	VipID *string `json:"vip_id"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsFsmEventDetails vs fsm event details
// swagger:model VsFsmEventDetails
type VsFsmEventDetails struct {

	// vip_id of VsFsmEventDetails.
	VipID *string `json:"vip_id,omitempty"`

	// Placeholder for description of property vs_rt of obj type VsFsmEventDetails field type str  type object
	VsRt *VirtualServiceRuntime `json:"vs_rt,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

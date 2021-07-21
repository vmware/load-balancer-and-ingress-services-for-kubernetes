// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsScaleInEventDetails vs scale in event details
// swagger:model VsScaleInEventDetails
type VsScaleInEventDetails struct {

	// error_message of VsScaleInEventDetails.
	ErrorMessage *string `json:"error_message,omitempty"`

	// ip of VsScaleInEventDetails.
	IP *string `json:"ip,omitempty"`

	// ip6 of VsScaleInEventDetails.
	Ip6 *string `json:"ip6,omitempty"`

	// Number of rpc_status.
	RPCStatus *int64 `json:"rpc_status,omitempty"`

	// Placeholder for description of property scale_status of obj type VsScaleInEventDetails field type str  type object
	ScaleStatus *ScaleStatus `json:"scale_status,omitempty"`

	// Placeholder for description of property se_assigned of obj type VsScaleInEventDetails field type str  type object
	SeAssigned []*VipSeAssigned `json:"se_assigned,omitempty"`

	// Placeholder for description of property se_requested of obj type VsScaleInEventDetails field type str  type object
	SeRequested *VirtualServiceResource `json:"se_requested,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RebalanceScaleinEventDetails rebalance scalein event details
// swagger:model RebalanceScaleinEventDetails
type RebalanceScaleinEventDetails struct {

	// Placeholder for description of property scalein_params of obj type RebalanceScaleinEventDetails field type str  type object
	ScaleinParams *VsScaleinParams `json:"scalein_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

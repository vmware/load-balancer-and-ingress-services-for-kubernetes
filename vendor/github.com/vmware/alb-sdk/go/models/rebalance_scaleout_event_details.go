// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RebalanceScaleoutEventDetails rebalance scaleout event details
// swagger:model RebalanceScaleoutEventDetails
type RebalanceScaleoutEventDetails struct {

	// Placeholder for description of property scaleout_params of obj type RebalanceScaleoutEventDetails field type str  type object
	ScaleoutParams *VsScaleoutParams `json:"scaleout_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

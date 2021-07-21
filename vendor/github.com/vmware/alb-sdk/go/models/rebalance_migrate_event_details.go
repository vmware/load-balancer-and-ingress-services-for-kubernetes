// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RebalanceMigrateEventDetails rebalance migrate event details
// swagger:model RebalanceMigrateEventDetails
type RebalanceMigrateEventDetails struct {

	// Placeholder for description of property migrate_params of obj type RebalanceMigrateEventDetails field type str  type object
	MigrateParams *VsMigrateParams `json:"migrate_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

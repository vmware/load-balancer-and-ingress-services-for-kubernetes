// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DisableSeMigrateEventDetails disable se migrate event details
// swagger:model DisableSeMigrateEventDetails
type DisableSeMigrateEventDetails struct {

	// Placeholder for description of property migrate_params of obj type DisableSeMigrateEventDetails field type str  type object
	MigrateParams *VsMigrateParams `json:"migrate_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

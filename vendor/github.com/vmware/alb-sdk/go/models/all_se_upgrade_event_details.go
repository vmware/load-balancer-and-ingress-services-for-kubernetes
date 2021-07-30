// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AllSeUpgradeEventDetails all se upgrade event details
// swagger:model AllSeUpgradeEventDetails
type AllSeUpgradeEventDetails struct {

	// notes of AllSeUpgradeEventDetails.
	Notes []string `json:"notes,omitempty"`

	// Number of num_se.
	// Required: true
	NumSe *int32 `json:"num_se"`

	// Number of num_vs.
	NumVs *int32 `json:"num_vs,omitempty"`

	// Placeholder for description of property request of obj type AllSeUpgradeEventDetails field type str  type object
	Request *SeUpgradeParams `json:"request,omitempty"`
}

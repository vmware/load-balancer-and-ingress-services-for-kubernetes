// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VinfraMgmtNwChangeDetails vinfra mgmt nw change details
// swagger:model VinfraMgmtNwChangeDetails
type VinfraMgmtNwChangeDetails struct {

	// existing_nw of VinfraMgmtNwChangeDetails.
	// Required: true
	ExistingNw *string `json:"existing_nw"`

	// new_nw of VinfraMgmtNwChangeDetails.
	// Required: true
	NewNw *string `json:"new_nw"`

	// vcenter of VinfraMgmtNwChangeDetails.
	// Required: true
	Vcenter *string `json:"vcenter"`
}

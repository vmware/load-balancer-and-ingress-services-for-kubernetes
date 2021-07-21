// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VlanRange vlan range
// swagger:model VlanRange
type VlanRange struct {

	// Number of end.
	// Required: true
	End *int32 `json:"end"`

	// Number of start.
	// Required: true
	Start *int32 `json:"start"`
}

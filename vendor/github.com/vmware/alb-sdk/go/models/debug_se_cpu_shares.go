// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugSeCPUShares debug se Cpu shares
// swagger:model DebugSeCpuShares
type DebugSeCPUShares struct {

	// Number of cpu.
	// Required: true
	CPU *int32 `json:"cpu"`

	// Number of shares.
	// Required: true
	Shares *int32 `json:"shares"`
}

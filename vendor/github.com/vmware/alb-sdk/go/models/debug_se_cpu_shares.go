package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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

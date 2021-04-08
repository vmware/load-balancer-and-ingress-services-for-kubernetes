package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MemoryBalancerInfo memory balancer info
// swagger:model MemoryBalancerInfo
type MemoryBalancerInfo struct {

	// Child Process information.
	Child []*ChildProcessInfo `json:"child,omitempty"`

	// Controller memory.
	ControllerMemory *int32 `json:"controller_memory,omitempty"`

	// Limit on the memory (in MB) for the Process.
	Limit *int32 `json:"limit,omitempty"`

	// Amount of memory (in MB) used by the Process.
	MemoryUsed *int32 `json:"memory_used,omitempty"`

	// PID of the Process.
	Pid *int32 `json:"pid,omitempty"`

	// Name of the Process.
	Process *string `json:"process,omitempty"`
}

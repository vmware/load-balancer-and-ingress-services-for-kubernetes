package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ChildProcessInfo child process info
// swagger:model ChildProcessInfo
type ChildProcessInfo struct {

	// Amount of memory (in MB) used by the sub process.
	Memory *int32 `json:"memory,omitempty"`

	// Process Id of the sub process.
	Pid *int32 `json:"pid,omitempty"`
}

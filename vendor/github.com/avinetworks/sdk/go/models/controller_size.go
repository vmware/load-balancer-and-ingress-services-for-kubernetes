package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerSize controller size
// swagger:model ControllerSize
type ControllerSize struct {

	// Controller flavor (E/S/M/L) for this controller size. Enum options - CONTROLLER_ESSENTIALS, CONTROLLER_SMALL, CONTROLLER_MEDIUM, CONTROLLER_LARGE. Field introduced in 20.1.1.
	Flavor *string `json:"flavor,omitempty"`

	// Minimum number of cpu cores required. Field introduced in 20.1.1.
	MinCpus *int32 `json:"min_cpus,omitempty"`

	// Minimum memory required. Field introduced in 20.1.1. Unit is GB.
	MinMemory *int32 `json:"min_memory,omitempty"`
}

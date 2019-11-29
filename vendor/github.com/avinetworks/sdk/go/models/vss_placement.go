package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VssPlacement vss placement
// swagger:model VssPlacement
type VssPlacement struct {

	// Degree of core non-affinity for VS placement. Allowed values are 1-256. Field introduced in 17.2.5.
	CoreNonaffinity *int32 `json:"core_nonaffinity,omitempty"`

	// Number of sub-cores that comprise a CPU core. Allowed values are 1-128. Field introduced in 17.2.5.
	NumSubcores *int32 `json:"num_subcores,omitempty"`
}

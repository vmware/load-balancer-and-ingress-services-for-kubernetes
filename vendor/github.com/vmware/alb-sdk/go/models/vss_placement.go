// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VssPlacement vss placement
// swagger:model VssPlacement
type VssPlacement struct {

	// Degree of core non-affinity for VS placement. Allowed values are 1-256. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CoreNonaffinity *uint32 `json:"core_nonaffinity,omitempty"`

	// Number of sub-cores that comprise a CPU core. Allowed values are 1-128. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSubcores *uint32 `json:"num_subcores,omitempty"`
}

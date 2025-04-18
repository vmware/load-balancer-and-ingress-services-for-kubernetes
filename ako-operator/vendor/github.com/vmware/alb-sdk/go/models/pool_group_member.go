// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolGroupMember pool group member
// swagger:model PoolGroupMember
type PoolGroupMember struct {

	// Pool deployment state used with the PG deployment policy. Enum options - EVALUATION_IN_PROGRESS, IN_SERVICE, OUT_OF_SERVICE, EVALUATION_FAILED. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeploymentState *string `json:"deployment_state,omitempty"`

	// UUID of the pool. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// All pools with same priority are treated similarly in a pool group. The higher the number, the higher the priority.A pool with a higher priority is selected, as long as the pool is eligible or an explicit policy chooses a different pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PriorityLabel *string `json:"priority_label,omitempty"`

	// Ratio of selecting eligible pools in the pool group. . Allowed values are 1-1000. Special values are 0 - Do not select this pool for new connections. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1), Basic edition(Allowed values- 1), Enterprise with Cloud Services edition.
	Ratio *uint32 `json:"ratio,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolGroupMember pool group member
// swagger:model PoolGroupMember
type PoolGroupMember struct {

	// Pool deployment state used with the PG deployment policy. Enum options - EVALUATION_IN_PROGRESS, IN_SERVICE, OUT_OF_SERVICE, EVALUATION_FAILED.
	DeploymentState *string `json:"deployment_state,omitempty"`

	// UUID of the pool. It is a reference to an object of type Pool.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// All pools with same label are treated similarly in a pool group. A pool with a higher priority is selected, as long as the pool is eligible or an explicit policy chooses a different pool.
	PriorityLabel *string `json:"priority_label,omitempty"`

	// Ratio of selecting eligible pools in the pool group. . Allowed values are 1-1000. Special values are 0 - 'Do not select this pool for new connections'.
	Ratio *int32 `json:"ratio,omitempty"`
}

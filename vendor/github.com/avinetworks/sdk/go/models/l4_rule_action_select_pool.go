package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4RuleActionSelectPool l4 rule action select pool
// swagger:model L4RuleActionSelectPool
type L4RuleActionSelectPool struct {

	// Indicates action to take on rule match. Enum options - L4_RULE_ACTION_SELECT_POOL, L4_RULE_ACTION_SELECT_POOLGROUP. Field introduced in 17.2.7.
	// Required: true
	ActionType *string `json:"action_type"`

	// ID of the pool group to serve the request. It is a reference to an object of type PoolGroup. Field introduced in 17.2.7.
	PoolGroupRef *string `json:"pool_group_ref,omitempty"`

	// ID of the pool of servers to serve the request. It is a reference to an object of type Pool. Field introduced in 17.2.7.
	PoolRef *string `json:"pool_ref,omitempty"`
}

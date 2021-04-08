package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4RuleAction l4 rule action
// swagger:model L4RuleAction
type L4RuleAction struct {

	// Indicates pool or pool-group selection on rule match. Field introduced in 17.2.7.
	SelectPool *L4RuleActionSelectPool `json:"select_pool,omitempty"`
}

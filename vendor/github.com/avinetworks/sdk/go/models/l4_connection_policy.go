package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4ConnectionPolicy l4 connection policy
// swagger:model L4ConnectionPolicy
type L4ConnectionPolicy struct {

	// Rules to apply when a new transport connection is setup. Field introduced in 17.2.7.
	Rules []*L4Rule `json:"rules,omitempty"`
}

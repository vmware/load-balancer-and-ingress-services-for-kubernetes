package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RoleFilterMatchLabel role filter match label
// swagger:model RoleFilterMatchLabel
type RoleFilterMatchLabel struct {

	// Key for filter match. Field introduced in 20.1.3.
	// Required: true
	Key *string `json:"key"`

	// Values for filter match. Multiple values will be evaluated as OR. Example  key = value1 OR key = value2. Behavior for match is key = * if this field is empty. Field introduced in 20.1.3.
	Values []string `json:"values,omitempty"`
}

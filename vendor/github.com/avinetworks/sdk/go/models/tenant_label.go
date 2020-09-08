package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TenantLabel tenant label
// swagger:model TenantLabel
type TenantLabel struct {

	// Label key string. Field introduced in 20.1.3.
	// Required: true
	Key *string `json:"key"`

	// Label value string. Field introduced in 20.1.3.
	Value *string `json:"value,omitempty"`
}

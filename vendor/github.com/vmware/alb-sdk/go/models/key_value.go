package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// KeyValue key value
// swagger:model KeyValue
type KeyValue struct {

	// Key.
	// Required: true
	Key *string `json:"key"`

	// Value.
	Value *string `json:"value,omitempty"`
}

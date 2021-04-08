package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// KeyValueTuple key value tuple
// swagger:model KeyValueTuple
type KeyValueTuple struct {

	// Key. Field introduced in 20.1.3.
	// Required: true
	Key *string `json:"key"`

	// Value. Field introduced in 20.1.3.
	Value *string `json:"value,omitempty"`
}

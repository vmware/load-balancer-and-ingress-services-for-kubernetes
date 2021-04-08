package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Selector selector
// swagger:model Selector
type Selector struct {

	// Labels as key value pairs to select on. Field introduced in 20.1.3. Minimum of 1 items required.
	Labels []*KeyValueTuple `json:"labels,omitempty"`

	// Selector type. Enum options - SELECTOR_IPAM. Field introduced in 20.1.3.
	// Required: true
	Type *string `json:"type"`
}

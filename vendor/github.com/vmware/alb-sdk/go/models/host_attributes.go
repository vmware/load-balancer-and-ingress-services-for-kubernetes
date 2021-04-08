package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HostAttributes host attributes
// swagger:model HostAttributes
type HostAttributes struct {

	// attr_key of HostAttributes.
	// Required: true
	AttrKey *string `json:"attr_key"`

	// attr_val of HostAttributes.
	AttrVal *string `json:"attr_val,omitempty"`
}

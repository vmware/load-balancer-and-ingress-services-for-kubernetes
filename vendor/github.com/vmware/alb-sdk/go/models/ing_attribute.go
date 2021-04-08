package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IngAttribute ing attribute
// swagger:model IngAttribute
type IngAttribute struct {

	// Attribute to match. Field introduced in 17.2.15, 18.1.5, 18.2.1.
	Attribute *string `json:"attribute,omitempty"`

	// Attribute value. If not set, match any value. Field introduced in 17.2.15, 18.1.5, 18.2.1.
	Value *string `json:"value,omitempty"`
}

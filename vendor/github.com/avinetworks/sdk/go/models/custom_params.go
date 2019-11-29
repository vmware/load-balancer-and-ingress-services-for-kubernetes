package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CustomParams custom params
// swagger:model CustomParams
type CustomParams struct {

	// Placeholder for description of property is_dynamic of obj type CustomParams field type str  type boolean
	IsDynamic *bool `json:"is_dynamic,omitempty"`

	// Placeholder for description of property is_sensitive of obj type CustomParams field type str  type boolean
	IsSensitive *bool `json:"is_sensitive,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// value of CustomParams.
	Value *string `json:"value,omitempty"`
}

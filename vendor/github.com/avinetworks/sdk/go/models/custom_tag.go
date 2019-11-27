package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CustomTag custom tag
// swagger:model CustomTag
type CustomTag struct {

	// tag_key of CustomTag.
	// Required: true
	TagKey *string `json:"tag_key"`

	// tag_val of CustomTag.
	TagVal *string `json:"tag_val,omitempty"`
}

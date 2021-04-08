package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Tag tag
// swagger:model Tag
type Tag struct {

	//  Enum options - AVI_DEFINED, USER_DEFINED, VCENTER_DEFINED.
	Type *string `json:"type,omitempty"`

	// value of Tag.
	// Required: true
	Value *string `json:"value"`
}

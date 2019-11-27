package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthTacacsPlusAttributeValuePair auth tacacs plus attribute value pair
// swagger:model AuthTacacsPlusAttributeValuePair
type AuthTacacsPlusAttributeValuePair struct {

	// mandatory.
	Mandatory *bool `json:"mandatory,omitempty"`

	// attribute name.
	Name *string `json:"name,omitempty"`

	// attribute value.
	Value *string `json:"value,omitempty"`
}

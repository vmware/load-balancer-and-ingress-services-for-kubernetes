package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SamlAttribute saml attribute
// swagger:model SamlAttribute
type SamlAttribute struct {

	// SAML Attribute name. Field introduced in 20.1.1.
	AttrName *string `json:"attr_name,omitempty"`

	// SAML Attribute values. Field introduced in 20.1.1.
	AttrValues []string `json:"attr_values,omitempty"`
}

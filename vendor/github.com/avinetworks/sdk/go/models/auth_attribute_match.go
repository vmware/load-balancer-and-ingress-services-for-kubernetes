package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthAttributeMatch auth attribute match
// swagger:model AuthAttributeMatch
type AuthAttributeMatch struct {

	// Attribute name whose values will be looked up in the access lists. Field introduced in 18.2.5.
	AttributeName *string `json:"attribute_name,omitempty"`

	// Attribute Values used to determine access when authentication applies. Field introduced in 18.2.5.
	AttributeValueList *StringMatch `json:"attribute_value_list,omitempty"`
}

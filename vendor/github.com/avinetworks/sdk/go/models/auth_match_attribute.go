package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthMatchAttribute auth match attribute
// swagger:model AuthMatchAttribute
type AuthMatchAttribute struct {

	// rule match criteria. Enum options - AUTH_MATCH_CONTAINS, AUTH_MATCH_DOES_NOT_CONTAIN, AUTH_MATCH_REGEX.
	Criteria *string `json:"criteria,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// values of AuthMatchAttribute.
	Values []string `json:"values,omitempty"`
}

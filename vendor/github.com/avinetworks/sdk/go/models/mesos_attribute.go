package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MesosAttribute mesos attribute
// swagger:model MesosAttribute
type MesosAttribute struct {

	// Attribute to match.
	// Required: true
	Attribute *string `json:"attribute"`

	// Attribute value. If not set, match any value.
	Value *string `json:"value,omitempty"`
}

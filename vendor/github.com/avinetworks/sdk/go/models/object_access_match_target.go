package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ObjectAccessMatchTarget object access match target
// swagger:model ObjectAccessMatchTarget
type ObjectAccessMatchTarget struct {

	// Key of the label to be matched. Field introduced in 18.2.7, 20.1.1.
	// Required: true
	LabelKey *string `json:"label_key"`

	// Label values that result in a successful match. Field introduced in 18.2.7, 20.1.1.
	// Required: true
	LabelValues []string `json:"label_values,omitempty"`
}

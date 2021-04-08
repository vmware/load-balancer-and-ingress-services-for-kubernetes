package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GeoDBMapping geo d b mapping
// swagger:model GeoDBMapping
type GeoDBMapping struct {

	// Description of the mapping. Field introduced in 21.1.1.
	Description *string `json:"description,omitempty"`

	// The set of mapping elements. Field introduced in 21.1.1.
	// Required: true
	Elements []*GeoDBMappingElement `json:"elements,omitempty"`

	// The unique name of the user mapping. Field introduced in 21.1.1.
	// Required: true
	Name *string `json:"name"`
}

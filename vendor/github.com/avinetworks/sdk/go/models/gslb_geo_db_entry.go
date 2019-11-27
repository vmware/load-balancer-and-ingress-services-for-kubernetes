package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbGeoDbEntry gslb geo db entry
// swagger:model GslbGeoDbEntry
type GslbGeoDbEntry struct {

	// This field describes the GeoDb file. Field introduced in 17.1.1.
	// Required: true
	File *GslbGeoDbFile `json:"file"`

	// Priority of this geodb entry. This value should be unique in a repeated list of geodb entries.  Higher the value, then greater is the priority.  . Allowed values are 1-100. Field introduced in 17.1.1.
	Priority *int32 `json:"priority,omitempty"`
}

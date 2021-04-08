package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GeoLocation geo location
// swagger:model GeoLocation
type GeoLocation struct {

	// Latitude of the location. This is represented as degrees.minutes. The range is from -90.0 (south) to +90.0 (north). Allowed values are -90.0-+90.0. Field introduced in 17.1.1.
	Latitude *float32 `json:"latitude,omitempty"`

	// Longitude of the location. This is represented as degrees.minutes. The range is from -180.0 (west) to +180.0 (east). Allowed values are -180.0-+180.0. Field introduced in 17.1.1.
	Longitude *float32 `json:"longitude,omitempty"`

	// Location name in the format Country/State/City. Field introduced in 17.1.1.
	Name *string `json:"name,omitempty"`

	// Location tag *string - example  USEast. Field introduced in 17.1.1.
	Tag *string `json:"tag,omitempty"`
}

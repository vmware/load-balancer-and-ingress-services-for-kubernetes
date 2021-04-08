package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbGeoLocation gslb geo location
// swagger:model GslbGeoLocation
type GslbGeoLocation struct {

	// Geographic location of the site. Field introduced in 17.1.1.
	Location *GeoLocation `json:"location,omitempty"`

	// This field describes the source of the GeoLocation. . Enum options - GSLB_LOCATION_SRC_USER_CONFIGURED, GSLB_LOCATION_SRC_INHERIT_FROM_SITE, GSLB_LOCATION_SRC_FROM_GEODB. Field introduced in 17.1.1.
	// Required: true
	Source *string `json:"source"`
}

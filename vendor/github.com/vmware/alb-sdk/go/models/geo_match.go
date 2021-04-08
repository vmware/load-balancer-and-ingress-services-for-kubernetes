package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GeoMatch geo match
// swagger:model GeoMatch
type GeoMatch struct {

	// The Geo data type to match on. Enum options - ATTRIBUTE_IP_PREFIX, ATTRIBUTE_COUNTRY_CODE, ATTRIBUTE_COUNTRY_NAME, ATTRIBUTE_CONTINENT_CODE, ATTRIBUTE_CONTINENT_NAME, ATTRIBUTE_REGION_NAME, ATTRIBUTE_CITY_NAME, ATTRIBUTE_ISP_NAME, ATTRIBUTE_ORGANIZATION_NAME, ATTRIBUTE_AS_NUMBER, ATTRIBUTE_AS_NAME, ATTRIBUTE_LONGITUDE, ATTRIBUTE_LATITUDE, ATTRIBUTE_CUSTOM_1, ATTRIBUTE_CUSTOM_2, ATTRIBUTE_CUSTOM_3, ATTRIBUTE_CUSTOM_4, ATTRIBUTE_CUSTOM_5, ATTRIBUTE_CUSTOM_6, ATTRIBUTE_CUSTOM_7.... Field introduced in 21.1.1.
	// Required: true
	Attribute *string `json:"attribute"`

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 21.1.1.
	// Required: true
	MatchOperation *string `json:"match_operation"`

	// The values to match. Field introduced in 21.1.1.
	// Required: true
	Values []string `json:"values,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GeoDBMappingElement geo d b mapping element
// swagger:model GeoDBMappingElement
type GeoDBMappingElement struct {

	// The attribute to map. Enum options - ATTRIBUTE_IP_PREFIX, ATTRIBUTE_COUNTRY_CODE, ATTRIBUTE_COUNTRY_NAME, ATTRIBUTE_CONTINENT_CODE, ATTRIBUTE_CONTINENT_NAME, ATTRIBUTE_REGION_NAME, ATTRIBUTE_CITY_NAME, ATTRIBUTE_ISP_NAME, ATTRIBUTE_ORGANIZATION_NAME, ATTRIBUTE_AS_NUMBER, ATTRIBUTE_AS_NAME, ATTRIBUTE_LONGITUDE, ATTRIBUTE_LATITUDE, ATTRIBUTE_CUSTOM_1, ATTRIBUTE_CUSTOM_2, ATTRIBUTE_CUSTOM_3, ATTRIBUTE_CUSTOM_4, ATTRIBUTE_CUSTOM_5, ATTRIBUTE_CUSTOM_6, ATTRIBUTE_CUSTOM_7.... Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Attribute *string `json:"attribute"`

	// The values to map. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Values []string `json:"values,omitempty"`
}

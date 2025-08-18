// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GeoLocation geo location
// swagger:model GeoLocation
type GeoLocation struct {

	// Latitude of the location. This is represented as degrees.minutes. The range is from -90.0 (south) to +90.0 (north). Allowed values are -90.0-+90.0. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Latitude *float32 `json:"latitude,omitempty"`

	// Longitude of the location. This is represented as degrees.minutes. The range is from -180.0 (west) to +180.0 (east). Allowed values are -180.0-+180.0. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Longitude *float32 `json:"longitude,omitempty"`

	// Location name in the format Country/State/City. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Location tag *string - example  USEast. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tag *string `json:"tag,omitempty"`
}

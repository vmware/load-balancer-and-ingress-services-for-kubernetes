// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbGeoDbEntry gslb geo db entry
// swagger:model GslbGeoDbEntry
type GslbGeoDbEntry struct {

	// This field describes the GeoDb file. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	File *GslbGeoDbFile `json:"file"`

	// Priority of this geodb entry. This value should be unique in a repeated list of geodb entries.  Higher the value, then greater is the priority.  . Allowed values are 1-100. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Priority *uint32 `json:"priority,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbGeoDbFile gslb geo db file
// swagger:model GslbGeoDbFile
type GslbGeoDbFile struct {

	// File checksum is internally computed. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	Checksum *string `json:"checksum,omitempty"`

	// System internal identifier for the file. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	FileID *string `json:"file_id,omitempty"`

	// Geodb Filename in the Avi supported formats. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Filename *string `json:"filename,omitempty"`

	// This field indicates the file format. Enum options - GSLB_GEODB_FILE_FORMAT_AVI, GSLB_GEODB_FILE_FORMAT_MAXMIND_CITY, GSLB_GEODB_FILE_FORMAT_MAXMIND_CITY_V6, GSLB_GEODB_FILE_FORMAT_MAXMIND_CITY_V4_AND_V6, GSLB_GEODB_FILE_FORMAT_AVI_V6, GSLB_GEODB_FILE_FORMAT_AVI_V4_AND_V6. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Format *string `json:"format,omitempty"`

	// Internal timestamp associated with the file. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	Timestamp uint64 `json:"timestamp,omitempty"`
}

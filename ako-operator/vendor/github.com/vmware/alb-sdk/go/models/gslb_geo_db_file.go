// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbGeoDbFile gslb geo db file
// swagger:model GslbGeoDbFile
type GslbGeoDbFile struct {

	// This field indicates the checksum of the original file. The checksum is internally computed. It's value changes every time the file is uploaded/modified. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	Checksum *string `json:"checksum,omitempty"`

	// This field indicates the internal file used in the system. The user uploaded file will be retained while a corresponding internal file is generated to be consumed by various upstream (Other sites) and downstream (SEs) entities. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	FileID *string `json:"file_id,omitempty"`

	// This field indicates the checksum of the internal file. The checksum is internally computed. It's value changes every time the internal file is regenerated. The internal file is regenerated whenever the original file is uploaded to the controller. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	FileIDChecksum *string `json:"file_id_checksum,omitempty"`

	// Geodb Filename in the Avi supported formats. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Filename *string `json:"filename,omitempty"`

	// This field indicates the file format. Enum options - GSLB_GEODB_FILE_FORMAT_AVI, GSLB_GEODB_FILE_FORMAT_MAXMIND_CITY, GSLB_GEODB_FILE_FORMAT_MAXMIND_CITY_V6, GSLB_GEODB_FILE_FORMAT_MAXMIND_CITY_V4_AND_V6, GSLB_GEODB_FILE_FORMAT_AVI_V6, GSLB_GEODB_FILE_FORMAT_AVI_V4_AND_V6. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Format *string `json:"format,omitempty"`

	// This field indicates the timestamp of when the file is associated to the GslbGeodbProfile. It is an internal generated timestamp. This value is a constant for the lifetime of the File and does not change every time the file is uploaded. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	Timestamp *uint64 `json:"timestamp,omitempty"`
}

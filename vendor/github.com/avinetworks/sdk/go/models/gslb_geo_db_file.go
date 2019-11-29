package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbGeoDbFile gslb geo db file
// swagger:model GslbGeoDbFile
type GslbGeoDbFile struct {

	// File checksum is internally computed. Field introduced in 17.1.1.
	// Read Only: true
	Checksum *string `json:"checksum,omitempty"`

	// System internal identifier for the file. Field introduced in 17.1.1.
	// Read Only: true
	FileID *string `json:"file_id,omitempty"`

	// Geodb Filename in the Avi supported formats. Field introduced in 17.1.1.
	Filename *string `json:"filename,omitempty"`

	// This field indicates the file format. Enum options - GSLB_GEODB_FILE_FORMAT_AVI, GSLB_GEODB_FILE_FORMAT_MAXMIND_CITY. Field introduced in 17.1.1.
	Format *string `json:"format,omitempty"`

	// Internal timestamp associated with the file. Field introduced in 17.1.1.
	// Read Only: true
	Timestamp *int64 `json:"timestamp,omitempty"`
}

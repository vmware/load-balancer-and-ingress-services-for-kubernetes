package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GeoDBFile geo d b file
// swagger:model GeoDBFile
type GeoDBFile struct {

	// If set to false, this file is ignored. Field introduced in 21.1.1.
	Enabled *bool `json:"enabled,omitempty"`

	// The file object that contains the geo data. Must be of type 'GeoDB'. It is a reference to an object of type FileObject. Field introduced in 21.1.1.
	// Required: true
	FileRef *string `json:"file_ref"`

	// Priority of the file. Field introduced in 21.1.1.
	// Required: true
	Index *int32 `json:"index"`

	// Name of the file. Field introduced in 21.1.1.
	// Required: true
	Name *string `json:"name"`

	// Source of the file data. Enum options - VENDOR_USER_DEFINED, VENDOR_AVI_DEFINED. Field introduced in 21.1.1.
	// Required: true
	Vendor *string `json:"vendor"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GeoDBFile geo d b file
// swagger:model GeoDBFile
type GeoDBFile struct {

	// If set to false, this file is ignored. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// The file object that contains the geo data. Must be of type 'GeoDB'. It is a reference to an object of type FileObject. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	FileRef *string `json:"file_ref"`

	// Priority of the file - larger number takes precedence. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	// Name of the file. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Source of the file data. Enum options - VENDOR_USER_DEFINED, VENDOR_AVI_DEFINED. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Vendor *string `json:"vendor"`
}

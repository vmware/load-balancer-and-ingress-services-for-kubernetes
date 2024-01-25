// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesFileDownloadMetadata a l b services file download metadata
// swagger:model ALBServicesFileDownloadMetadata
type ALBServicesFileDownloadMetadata struct {

	// Checksum of the file. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Checksum *string `json:"checksum,omitempty"`

	// Currently only MD5 checksum type is supported. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ChecksumType *string `json:"checksum_type,omitempty"`

	// Checksum size in bytes. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ChunkSize uint64 `json:"chunk_size,omitempty"`

	// Whether the file can be downloaded in parts or not. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	IsMultiPartDownload *bool `json:"is_multi_part_download"`

	// Sigend url of the file from pulse. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	SignedURL *string `json:"signed_url"`

	// Total size of the file in bytes. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	TotalSize *uint64 `json:"total_size"`
}

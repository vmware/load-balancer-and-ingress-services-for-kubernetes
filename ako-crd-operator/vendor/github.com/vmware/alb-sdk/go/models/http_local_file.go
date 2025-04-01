// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPLocalFile HTTP local file
// swagger:model HTTPLocalFile
type HTTPLocalFile struct {

	// Mime-type of the content in the file. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ContentType *string `json:"content_type"`

	// File content to used in the local HTTP response body. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	FileContent *string `json:"file_content"`

	// File content length. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	FileLength uint32 `json:"file_length,omitempty"`
}

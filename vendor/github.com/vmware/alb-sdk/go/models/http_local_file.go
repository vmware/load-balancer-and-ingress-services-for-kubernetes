// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPLocalFile HTTP local file
// swagger:model HTTPLocalFile
type HTTPLocalFile struct {

	// Mime-type of the content in the file.
	// Required: true
	ContentType *string `json:"content_type"`

	// File content to used in the local HTTP response body.
	// Required: true
	FileContent *string `json:"file_content"`
}

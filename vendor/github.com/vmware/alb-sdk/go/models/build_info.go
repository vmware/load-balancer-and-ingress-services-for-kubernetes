// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BuildInfo build info
// swagger:model BuildInfo
type BuildInfo struct {

	// Build number for easy identification. Field introduced in 18.2.6.
	BuildNo *int32 `json:"build_no,omitempty"`

	// Date when the package created. Field introduced in 18.2.6.
	Date *string `json:"date,omitempty"`

	// Min version of the image. Field introduced in 18.2.6.
	MinVersion *string `json:"min_version,omitempty"`

	// Patch version of the image. Field introduced in 18.2.6.
	PatchVersion *string `json:"patch_version,omitempty"`

	// Product type. Field introduced in 18.2.6.
	Product *string `json:"product,omitempty"`

	// Product Name. Field introduced in 18.2.6.
	ProductName *string `json:"product_name,omitempty"`

	// Tag related to the package. Field introduced in 18.2.6.
	Tag *string `json:"tag,omitempty"`

	// Major version of the image. Field introduced in 18.2.6.
	Version *string `json:"version,omitempty"`
}

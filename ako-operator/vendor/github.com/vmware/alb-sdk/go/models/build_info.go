// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BuildInfo build info
// swagger:model BuildInfo
type BuildInfo struct {

	// Build number for easy identification. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BuildNo *int32 `json:"build_no,omitempty"`

	// Date when the package created. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Date *string `json:"date,omitempty"`

	// Min version of the image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinVersion *string `json:"min_version,omitempty"`

	// Patch version of the image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchVersion *string `json:"patch_version,omitempty"`

	// Product type. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Product *string `json:"product,omitempty"`

	// Product Name. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProductName *string `json:"product_name,omitempty"`

	// Remote reference of the container image. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RemoteImageRef *string `json:"remote_image_ref,omitempty"`

	// Tag related to the package. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tag *string `json:"tag,omitempty"`

	// Major version of the image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *string `json:"version,omitempty"`
}

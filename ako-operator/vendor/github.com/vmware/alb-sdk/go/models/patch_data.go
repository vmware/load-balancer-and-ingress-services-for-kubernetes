// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PatchData patch data
// swagger:model PatchData
type PatchData struct {

	// Image path of current patch image. . Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchImagePath *string `json:"patch_image_path,omitempty"`

	// Image uuid for identifying the patch. It is a reference to an object of type Image. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchImageRef *string `json:"patch_image_ref,omitempty"`

	// Patch version. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchVersion *string `json:"patch_version,omitempty"`
}

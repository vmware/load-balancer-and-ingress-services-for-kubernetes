// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FileReferenceMapping file reference mapping
// swagger:model FileReferenceMapping
type FileReferenceMapping struct {

	// Absolute file path corresponding to the reference. Supported parameters in file_path are {image_path}, {current_version} and {prev_version}. For example, {image_path}/{prev_version}/se_nsxt.ova would resolve to /vol/pkgs/30.1.1-9000-20230714.075215/se_nsxt.ova. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	FilePath *string `json:"file_path"`

	// Short named reference for file path. For example, SE_IMG. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Reference *string `json:"reference"`
}

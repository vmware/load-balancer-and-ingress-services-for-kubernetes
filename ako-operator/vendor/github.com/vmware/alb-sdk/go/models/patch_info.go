// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PatchInfo patch info
// swagger:model PatchInfo
type PatchInfo struct {

	// Patch type describes the controller or se patch type. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchType *string `json:"patch_type,omitempty"`

	// This variable tells whether reboot has to be performed. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reboot *bool `json:"reboot,omitempty"`

	// This variable is for full list of patch reboot details. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RebootList []*RebootData `json:"reboot_list,omitempty"`
}

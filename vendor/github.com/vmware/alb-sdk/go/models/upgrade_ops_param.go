// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UpgradeOpsParam upgrade ops param
// swagger:model UpgradeOpsParam
type UpgradeOpsParam struct {

	// Image uuid for identifying base image. It is a reference to an object of type Image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ImageRef *string `json:"image_ref,omitempty"`

	// Image uuid for identifying patch. It is a reference to an object of type Image. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PatchRef *string `json:"patch_ref,omitempty"`

	// This field identifies SE group options that need to be applied during the upgrade operations. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupOptions *SeGroupOptions `json:"se_group_options,omitempty"`

	// Apply options while resuming SE group upgrade operations. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupResumeOptions *SeGroupResumeOptions `json:"se_group_resume_options,omitempty"`
}

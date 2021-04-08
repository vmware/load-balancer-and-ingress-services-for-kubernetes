package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeOpsParam upgrade ops param
// swagger:model UpgradeOpsParam
type UpgradeOpsParam struct {

	// Image uuid for identifying base image. It is a reference to an object of type Image. Field introduced in 18.2.6.
	ImageRef *string `json:"image_ref,omitempty"`

	// Image uuid for identifying patch. It is a reference to an object of type Image. Field introduced in 18.2.6.
	PatchRef *string `json:"patch_ref,omitempty"`

	// This field identifies SE group options that need to be applied during the upgrade operations. Field introduced in 18.2.6.
	SeGroupOptions *SeGroupOptions `json:"se_group_options,omitempty"`

	// Apply options while resuming SE group upgrade operations. Field introduced in 18.2.6.
	SeGroupResumeOptions *SeGroupResumeOptions `json:"se_group_resume_options,omitempty"`
}

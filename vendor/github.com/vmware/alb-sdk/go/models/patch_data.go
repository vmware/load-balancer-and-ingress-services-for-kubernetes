package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PatchData patch data
// swagger:model PatchData
type PatchData struct {

	// Image path of current patch image. . Field introduced in 18.2.10, 20.1.1.
	PatchImagePath *string `json:"patch_image_path,omitempty"`

	// Image uuid for identifying the patch. It is a reference to an object of type Image. Field introduced in 18.2.8, 20.1.1.
	PatchImageRef *string `json:"patch_image_ref,omitempty"`

	// Patch version. Field introduced in 18.2.8, 20.1.1.
	PatchVersion *string `json:"patch_version,omitempty"`
}

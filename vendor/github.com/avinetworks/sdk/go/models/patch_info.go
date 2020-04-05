package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PatchInfo patch info
// swagger:model PatchInfo
type PatchInfo struct {

	// Patch type describes the controller or se patch type. Field introduced in 18.2.6.
	PatchType *string `json:"patch_type,omitempty"`

	// This variable tells whether reboot has to be performed. Field introduced in 18.2.6.
	Reboot *bool `json:"reboot,omitempty"`

	// This variable is for full list of patch reboot details. Field introduced in 18.2.8.
	RebootList []*RebootData `json:"reboot_list,omitempty"`
}

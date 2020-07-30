package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RebootData reboot data
// swagger:model RebootData
type RebootData struct {

	// Patch version for which reboot flag need to be computed. Field introduced in 18.2.8, 20.1.1.
	PatchVersion *string `json:"patch_version,omitempty"`

	// This variable tells whether reboot has to be performed. Field introduced in 18.2.8, 20.1.1.
	Reboot *bool `json:"reboot,omitempty"`
}

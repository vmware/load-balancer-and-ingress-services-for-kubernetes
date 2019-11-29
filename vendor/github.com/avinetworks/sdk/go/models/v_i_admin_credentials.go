package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIAdminCredentials v i admin credentials
// swagger:model VIAdminCredentials
type VIAdminCredentials struct {

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// password of VIAdminCredentials.
	Password *string `json:"password,omitempty"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`

	// vcenter_url of VIAdminCredentials.
	// Required: true
	VcenterURL *string `json:"vcenter_url"`

	// vi_mgr_token of VIAdminCredentials.
	ViMgrToken *string `json:"vi_mgr_token,omitempty"`
}

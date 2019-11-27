package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIdeleteSEReq v idelete s e req
// swagger:model VIDeleteSEReq
type VIdeleteSEReq struct {

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// Unique object identifier of segroup.
	SegroupUUID *string `json:"segroup_uuid,omitempty"`

	// Unique object identifier of sevm.
	// Required: true
	SevmUUID *string `json:"sevm_uuid"`

	// Placeholder for description of property vcenter_admin of obj type VIDeleteSEReq field type str  type object
	VcenterAdmin *VIAdminCredentials `json:"vcenter_admin,omitempty"`
}

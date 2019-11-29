package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VISetvNicNwReq v i setv nic nw req
// swagger:model VISetvNicNwReq
type VISetvNicNwReq struct {

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// Unique object identifier of dc.
	DcUUID *string `json:"dc_uuid,omitempty"`

	// rm_cookie of VISetvNicNwReq.
	RmCookie *string `json:"rm_cookie,omitempty"`

	// Unique object identifier of sevm.
	// Required: true
	SevmUUID *string `json:"sevm_uuid"`

	// Placeholder for description of property vcenter_admin of obj type VISetvNicNwReq field type str  type object
	VcenterAdmin *VIAdminCredentials `json:"vcenter_admin,omitempty"`

	// vcenter_sevm_mor of VISetvNicNwReq.
	VcenterSevmMor *string `json:"vcenter_sevm_mor,omitempty"`

	// Placeholder for description of property vcenter_vnic_info of obj type VISetvNicNwReq field type str  type object
	VcenterVnicInfo []*VIVMVnicInfo `json:"vcenter_vnic_info,omitempty"`
}

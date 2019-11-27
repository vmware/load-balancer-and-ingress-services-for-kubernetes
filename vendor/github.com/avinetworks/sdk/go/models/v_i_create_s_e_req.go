package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VICreateSEReq v i create s e req
// swagger:model VICreateSEReq
type VICreateSEReq struct {

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// cookie of VICreateSEReq.
	Cookie *string `json:"cookie,omitempty"`

	// Unique object identifier of se_grp.
	SeGrpUUID *string `json:"se_grp_uuid,omitempty"`

	// Placeholder for description of property se_params of obj type VICreateSEReq field type str  type object
	// Required: true
	SeParams *VISeVMOvaParams `json:"se_params"`

	// Unique object identifier of tenant.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// Placeholder for description of property vcenter_admin of obj type VICreateSEReq field type str  type object
	VcenterAdmin *VIAdminCredentials `json:"vcenter_admin,omitempty"`

	// vcenter_vnic_portgroups of VICreateSEReq.
	VcenterVnicPortgroups []string `json:"vcenter_vnic_portgroups,omitempty"`
}

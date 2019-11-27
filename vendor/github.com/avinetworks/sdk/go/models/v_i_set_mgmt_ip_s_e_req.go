package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VISetMgmtIPSEReq v i set mgmt Ip s e req
// swagger:model VISetMgmtIpSEReq
type VISetMgmtIPSEReq struct {

	// Placeholder for description of property admin of obj type VISetMgmtIpSEReq field type str  type object
	Admin *VIAdminCredentials `json:"admin,omitempty"`

	// Placeholder for description of property all_vnic_connected of obj type VISetMgmtIpSEReq field type str  type boolean
	AllVnicConnected *bool `json:"all_vnic_connected,omitempty"`

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// Unique object identifier of dc.
	DcUUID *string `json:"dc_uuid,omitempty"`

	// Placeholder for description of property ip_params of obj type VISetMgmtIpSEReq field type str  type object
	// Required: true
	IPParams *VISeVMIPConfParams `json:"ip_params"`

	// Placeholder for description of property power_on of obj type VISetMgmtIpSEReq field type str  type boolean
	PowerOn *bool `json:"power_on,omitempty"`

	// rm_cookie of VISetMgmtIpSEReq.
	RmCookie *string `json:"rm_cookie,omitempty"`

	// Unique object identifier of sevm.
	// Required: true
	SevmUUID *string `json:"sevm_uuid"`
}

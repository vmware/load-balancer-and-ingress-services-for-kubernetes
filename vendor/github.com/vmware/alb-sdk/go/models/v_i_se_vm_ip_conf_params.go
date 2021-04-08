package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VISeVMIPConfParams v i se Vm Ip conf params
// swagger:model VISeVmIpConfParams
type VISeVMIPConfParams struct {

	// default_gw of VISeVmIpConfParams.
	DefaultGw *string `json:"default_gw,omitempty"`

	// mgmt_ip_addr of VISeVmIpConfParams.
	MgmtIPAddr *string `json:"mgmt_ip_addr,omitempty"`

	//  Enum options - VNIC_IP_TYPE_DHCP, VNIC_IP_TYPE_STATIC.
	// Required: true
	MgmtIPType *string `json:"mgmt_ip_type"`

	// mgmt_net_mask of VISeVmIpConfParams.
	MgmtNetMask *string `json:"mgmt_net_mask,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIVMVnicInfo v i Vm vnic info
// swagger:model VIVmVnicInfo
type VIVMVnicInfo struct {

	// mac_addr of VIVmVnicInfo.
	// Required: true
	MacAddr *string `json:"mac_addr"`

	// vcenter_portgroup of VIVmVnicInfo.
	VcenterPortgroup *string `json:"vcenter_portgroup,omitempty"`

	//  Enum options - VNIC_VSWITCH, VNIC_DVS.
	VcenterVnicNw *string `json:"vcenter_vnic_nw,omitempty"`
}

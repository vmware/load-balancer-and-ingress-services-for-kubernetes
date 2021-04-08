package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIControllerVnicInfo v i controller vnic info
// swagger:model VIControllerVnicInfo
type VIControllerVnicInfo struct {

	// portgroup of VIControllerVnicInfo.
	// Required: true
	Portgroup *string `json:"portgroup"`

	// Placeholder for description of property vnic_ip of obj type VIControllerVnicInfo field type str  type object
	VnicIP []*VIGuestvNicIPAddr `json:"vnic_ip,omitempty"`
}

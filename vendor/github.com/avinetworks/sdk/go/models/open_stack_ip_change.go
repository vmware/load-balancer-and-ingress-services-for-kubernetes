package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenStackIPChange open stack Ip change
// swagger:model OpenStackIpChange
type OpenStackIPChange struct {

	// error_string of OpenStackIpChange.
	ErrorString *string `json:"error_string,omitempty"`

	// Placeholder for description of property ip of obj type OpenStackIpChange field type str  type object
	// Required: true
	IP *IPAddr `json:"ip"`

	// mac_addr of OpenStackIpChange.
	MacAddr *string `json:"mac_addr,omitempty"`

	// Unique object identifier of port.
	PortUUID *string `json:"port_uuid,omitempty"`

	// Unique object identifier of se_vm.
	SeVMUUID *string `json:"se_vm_uuid,omitempty"`
}

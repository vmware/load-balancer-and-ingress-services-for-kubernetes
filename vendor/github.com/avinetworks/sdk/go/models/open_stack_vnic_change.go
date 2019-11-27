package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenStackVnicChange open stack vnic change
// swagger:model OpenStackVnicChange
type OpenStackVnicChange struct {

	// error_string of OpenStackVnicChange.
	ErrorString *string `json:"error_string,omitempty"`

	// mac_addrs of OpenStackVnicChange.
	MacAddrs []string `json:"mac_addrs,omitempty"`

	// networks of OpenStackVnicChange.
	Networks []string `json:"networks,omitempty"`

	// Unique object identifier of se_vm.
	// Required: true
	SeVMUUID *string `json:"se_vm_uuid"`
}

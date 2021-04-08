package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RmModifyVnic rm modify vnic
// swagger:model RmModifyVnic
type RmModifyVnic struct {

	// mac_addr of RmModifyVnic.
	MacAddr *string `json:"mac_addr,omitempty"`

	// network_name of RmModifyVnic.
	NetworkName *string `json:"network_name,omitempty"`

	// Unique object identifier of network.
	NetworkUUID *string `json:"network_uuid,omitempty"`
}

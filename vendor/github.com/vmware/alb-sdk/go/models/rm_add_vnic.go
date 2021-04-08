package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RmAddVnic rm add vnic
// swagger:model RmAddVnic
type RmAddVnic struct {

	// network_name of RmAddVnic.
	NetworkName *string `json:"network_name,omitempty"`

	// Unique object identifier of network.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	// subnet of RmAddVnic.
	Subnet *string `json:"subnet,omitempty"`
}

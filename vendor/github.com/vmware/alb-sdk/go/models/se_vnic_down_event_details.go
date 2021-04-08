package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeVnicDownEventDetails se vnic down event details
// swagger:model SeVnicDownEventDetails
type SeVnicDownEventDetails struct {

	// Vnic name.
	IfName *string `json:"if_name,omitempty"`

	// Vnic linux name.
	LinuxName *string `json:"linux_name,omitempty"`

	// Mac Address.
	Mac *string `json:"mac,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`
}

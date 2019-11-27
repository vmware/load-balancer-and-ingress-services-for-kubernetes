package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeIPAddedEventDetails se Ip added event details
// swagger:model SeIpAddedEventDetails
type SeIPAddedEventDetails struct {

	// Vnic name.
	IfName *string `json:"if_name,omitempty"`

	// IP added.
	IP *string `json:"ip,omitempty"`

	// Vnic linux name.
	LinuxName *string `json:"linux_name,omitempty"`

	// Mac Address.
	Mac *string `json:"mac,omitempty"`

	// Mask .
	Mask *int32 `json:"mask,omitempty"`

	// DCHP or Static.
	Mode *string `json:"mode,omitempty"`

	// Network UUID.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	// Namespace.
	Ns *string `json:"ns,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`
}

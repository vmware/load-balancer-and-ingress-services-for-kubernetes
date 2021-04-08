package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeIpfailureEventDetails se ipfailure event details
// swagger:model SeIpfailureEventDetails
type SeIpfailureEventDetails struct {

	// Mac Address.
	Mac *string `json:"mac,omitempty"`

	// Network UUID.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`

	// Vnic name.
	VnicName *string `json:"vnic_name,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RmSeIPFailEventDetails rm se Ip fail event details
// swagger:model RmSeIpFailEventDetails
type RmSeIPFailEventDetails struct {

	// host_name of RmSeIpFailEventDetails.
	HostName *string `json:"host_name,omitempty"`

	// Placeholder for description of property networks of obj type RmSeIpFailEventDetails field type str  type object
	Networks []*RmAddVnic `json:"networks,omitempty"`

	// reason of RmSeIpFailEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_name of RmSeIpFailEventDetails.
	SeName *string `json:"se_name,omitempty"`
}

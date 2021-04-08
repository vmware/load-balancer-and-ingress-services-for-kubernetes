package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RmAddNetworksEventDetails rm add networks event details
// swagger:model RmAddNetworksEventDetails
type RmAddNetworksEventDetails struct {

	// Placeholder for description of property networks of obj type RmAddNetworksEventDetails field type str  type object
	Networks []*RmAddVnic `json:"networks,omitempty"`

	// reason of RmAddNetworksEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_name of RmAddNetworksEventDetails.
	SeName *string `json:"se_name,omitempty"`

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// vs_name of RmAddNetworksEventDetails.
	VsName []string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID []string `json:"vs_uuid,omitempty"`
}

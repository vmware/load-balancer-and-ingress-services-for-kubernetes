package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RmModifyNetworksEventDetails rm modify networks event details
// swagger:model RmModifyNetworksEventDetails
type RmModifyNetworksEventDetails struct {

	// Placeholder for description of property networks of obj type RmModifyNetworksEventDetails field type str  type object
	Networks []*RmModifyVnic `json:"networks,omitempty"`

	// reason of RmModifyNetworksEventDetails.
	Reason *string `json:"reason,omitempty"`

	// se_name of RmModifyNetworksEventDetails.
	SeName *string `json:"se_name,omitempty"`

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// vs_name of RmModifyNetworksEventDetails.
	VsName []string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID []string `json:"vs_uuid,omitempty"`
}

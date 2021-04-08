package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsSeVnic vs se vnic
// swagger:model VsSeVnic
type VsSeVnic struct {

	// lif of VsSeVnic.
	Lif *string `json:"lif,omitempty"`

	// mac of VsSeVnic.
	// Required: true
	Mac *string `json:"mac"`

	//  Enum options - VNIC_TYPE_FE, VNIC_TYPE_BE.
	// Required: true
	Type *string `json:"type"`
}

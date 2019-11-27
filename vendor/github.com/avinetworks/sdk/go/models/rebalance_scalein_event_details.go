package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RebalanceScaleinEventDetails rebalance scalein event details
// swagger:model RebalanceScaleinEventDetails
type RebalanceScaleinEventDetails struct {

	// Placeholder for description of property scalein_params of obj type RebalanceScaleinEventDetails field type str  type object
	ScaleinParams *VsScaleinParams `json:"scalein_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

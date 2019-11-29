package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeScaleinEventDetails se upgrade scalein event details
// swagger:model SeUpgradeScaleinEventDetails
type SeUpgradeScaleinEventDetails struct {

	// Placeholder for description of property scalein_params of obj type SeUpgradeScaleinEventDetails field type str  type object
	ScaleinParams *VsScaleinParams `json:"scalein_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

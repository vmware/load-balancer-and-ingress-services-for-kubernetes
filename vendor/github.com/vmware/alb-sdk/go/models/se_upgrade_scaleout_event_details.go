package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeScaleoutEventDetails se upgrade scaleout event details
// swagger:model SeUpgradeScaleoutEventDetails
type SeUpgradeScaleoutEventDetails struct {

	// Placeholder for description of property scaleout_params of obj type SeUpgradeScaleoutEventDetails field type str  type object
	ScaleoutParams *VsScaleoutParams `json:"scaleout_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeEventDetails se upgrade event details
// swagger:model SeUpgradeEventDetails
type SeUpgradeEventDetails struct {

	// notes of SeUpgradeEventDetails.
	Notes []string `json:"notes,omitempty"`

	// Number of num_vs.
	NumVs *int32 `json:"num_vs,omitempty"`

	// Unique object identifier of se_grp.
	SeGrpUUID *string `json:"se_grp_uuid,omitempty"`

	// Unique object identifier of se.
	// Required: true
	SeUUID *string `json:"se_uuid"`
}

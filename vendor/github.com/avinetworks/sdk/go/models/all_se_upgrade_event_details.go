package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AllSeUpgradeEventDetails all se upgrade event details
// swagger:model AllSeUpgradeEventDetails
type AllSeUpgradeEventDetails struct {

	// notes of AllSeUpgradeEventDetails.
	Notes []string `json:"notes,omitempty"`

	// Number of num_se.
	// Required: true
	NumSe *int32 `json:"num_se"`

	// Number of num_vs.
	NumVs *int32 `json:"num_vs,omitempty"`

	// Placeholder for description of property request of obj type AllSeUpgradeEventDetails field type str  type object
	Request *SeUpgradeParams `json:"request,omitempty"`
}

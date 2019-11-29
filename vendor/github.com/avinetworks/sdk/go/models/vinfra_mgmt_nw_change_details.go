package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraMgmtNwChangeDetails vinfra mgmt nw change details
// swagger:model VinfraMgmtNwChangeDetails
type VinfraMgmtNwChangeDetails struct {

	// existing_nw of VinfraMgmtNwChangeDetails.
	// Required: true
	ExistingNw *string `json:"existing_nw"`

	// new_nw of VinfraMgmtNwChangeDetails.
	// Required: true
	NewNw *string `json:"new_nw"`

	// vcenter of VinfraMgmtNwChangeDetails.
	// Required: true
	Vcenter *string `json:"vcenter"`
}

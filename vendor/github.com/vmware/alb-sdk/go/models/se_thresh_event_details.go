package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeThreshEventDetails se thresh event details
// swagger:model SeThreshEventDetails
type SeThreshEventDetails struct {

	// Number of curr_value.
	// Required: true
	CurrValue *int64 `json:"curr_value"`

	// Number of thresh.
	// Required: true
	Thresh *int64 `json:"thresh"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VlanRange vlan range
// swagger:model VlanRange
type VlanRange struct {

	// Number of end.
	// Required: true
	End *int32 `json:"end"`

	// Number of start.
	// Required: true
	Start *int32 `json:"start"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PortRange port range
// swagger:model PortRange
type PortRange struct {

	// TCP/UDP port range end (inclusive). Allowed values are 1-65535.
	// Required: true
	End *int32 `json:"end"`

	// TCP/UDP port range start (inclusive). Allowed values are 1-65535.
	// Required: true
	Start *int32 `json:"start"`
}

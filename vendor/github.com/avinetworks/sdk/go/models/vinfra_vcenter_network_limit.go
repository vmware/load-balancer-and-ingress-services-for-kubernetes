package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraVcenterNetworkLimit vinfra vcenter network limit
// swagger:model VinfraVcenterNetworkLimit
type VinfraVcenterNetworkLimit struct {

	// additional_reason of VinfraVcenterNetworkLimit.
	// Required: true
	AdditionalReason *string `json:"additional_reason"`

	// Number of current.
	// Required: true
	Current *int64 `json:"current"`

	// Number of limit.
	// Required: true
	Limit *int64 `json:"limit"`
}

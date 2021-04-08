package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraVcenterDiscoveryFailure vinfra vcenter discovery failure
// swagger:model VinfraVcenterDiscoveryFailure
type VinfraVcenterDiscoveryFailure struct {

	// state of VinfraVcenterDiscoveryFailure.
	// Required: true
	State *string `json:"state"`

	// vcenter of VinfraVcenterDiscoveryFailure.
	// Required: true
	Vcenter *string `json:"vcenter"`
}

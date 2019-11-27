package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraVMDetails vinfra Vm details
// swagger:model VinfraVmDetails
type VinfraVMDetails struct {

	// datacenter of VinfraVmDetails.
	Datacenter *string `json:"datacenter,omitempty"`

	// host of VinfraVmDetails.
	Host *string `json:"host,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`
}

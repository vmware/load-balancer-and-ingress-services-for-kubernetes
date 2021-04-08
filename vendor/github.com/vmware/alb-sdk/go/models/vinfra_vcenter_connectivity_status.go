package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VinfraVcenterConnectivityStatus vinfra vcenter connectivity status
// swagger:model VinfraVcenterConnectivityStatus
type VinfraVcenterConnectivityStatus struct {

	// cloud of VinfraVcenterConnectivityStatus.
	// Required: true
	Cloud *string `json:"cloud"`

	// datacenter of VinfraVcenterConnectivityStatus.
	// Required: true
	Datacenter *string `json:"datacenter"`

	// vcenter of VinfraVcenterConnectivityStatus.
	// Required: true
	Vcenter *string `json:"vcenter"`
}

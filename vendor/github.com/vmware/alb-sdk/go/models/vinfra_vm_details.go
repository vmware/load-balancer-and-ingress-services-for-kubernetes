// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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

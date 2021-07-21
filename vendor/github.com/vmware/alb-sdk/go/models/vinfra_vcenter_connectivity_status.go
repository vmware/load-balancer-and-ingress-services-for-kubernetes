// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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

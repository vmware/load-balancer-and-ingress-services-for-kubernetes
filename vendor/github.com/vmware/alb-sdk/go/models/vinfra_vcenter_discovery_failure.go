// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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

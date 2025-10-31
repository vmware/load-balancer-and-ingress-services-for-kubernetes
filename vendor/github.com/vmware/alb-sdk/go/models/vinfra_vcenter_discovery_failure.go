// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VinfraVcenterDiscoveryFailure vinfra vcenter discovery failure
// swagger:model VinfraVcenterDiscoveryFailure
type VinfraVcenterDiscoveryFailure struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	State *string `json:"state"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Vcenter *string `json:"vcenter"`
}

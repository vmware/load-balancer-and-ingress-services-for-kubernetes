// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VinfraVcenterConnectivityStatus vinfra vcenter connectivity status
// swagger:model VinfraVcenterConnectivityStatus
type VinfraVcenterConnectivityStatus struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Cloud *string `json:"cloud"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Datacenter *string `json:"datacenter"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Vcenter *string `json:"vcenter"`
}

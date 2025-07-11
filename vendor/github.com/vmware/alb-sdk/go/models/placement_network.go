// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PlacementNetwork placement network
// swagger:model PlacementNetwork
type PlacementNetwork struct {

	//  It is a reference to an object of type Network. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NetworkRef *string `json:"network_ref"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Subnet *IPAddrPrefix `json:"subnet"`
}

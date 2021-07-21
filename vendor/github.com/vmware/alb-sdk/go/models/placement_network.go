// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PlacementNetwork placement network
// swagger:model PlacementNetwork
type PlacementNetwork struct {

	//  It is a reference to an object of type Network.
	// Required: true
	NetworkRef *string `json:"network_ref"`

	// Placeholder for description of property subnet of obj type PlacementNetwork field type str  type object
	// Required: true
	Subnet *IPAddrPrefix `json:"subnet"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DiscoveredNetwork discovered network
// swagger:model DiscoveredNetwork
type DiscoveredNetwork struct {

	// Discovered network for this IP. It is a reference to an object of type Network.
	// Required: true
	NetworkRef *string `json:"network_ref"`

	// Discovered subnet for this IP.
	Subnet []*IPAddrPrefix `json:"subnet,omitempty"`

	// Discovered IPv6 subnet for this IP. Field introduced in 18.1.1.
	Subnet6 []*IPAddrPrefix `json:"subnet6,omitempty"`
}

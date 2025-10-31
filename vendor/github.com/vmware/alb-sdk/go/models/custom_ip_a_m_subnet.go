// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CustomIPAMSubnet custom ipam subnet
// swagger:model CustomIpamSubnet
type CustomIPAMSubnet struct {

	// Network to use for Custom IPAM IP allocation. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	NetworkID *string `json:"network_id"`

	// IPv4 subnet to use for Custom IPAM IP allocation. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// IPv6 subnet to use for Custom IPAM IP allocation. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Subnet6 *IPAddrPrefix `json:"subnet6,omitempty"`
}

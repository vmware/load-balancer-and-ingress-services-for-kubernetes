// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Subnet subnet
// swagger:model Subnet
type Subnet struct {

	// Specify an IP subnet prefix for this Network.
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// Static IP ranges for this subnet. Field introduced in 20.1.3.
	StaticIPRanges []*StaticIPRange `json:"static_ip_ranges,omitempty"`

	// Use static_ip_ranges. Field deprecated in 20.1.3.
	StaticIps []*IPAddr `json:"static_ips,omitempty"`

	// Use static_ip_ranges. Field deprecated in 20.1.3.
	StaticRanges []*IPAddrRange `json:"static_ranges,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NatAddrInfo nat addr info
// swagger:model NatAddrInfo
type NatAddrInfo struct {

	// Nat IP address. Field introduced in 18.2.3.
	NatIP *IPAddr `json:"nat_ip,omitempty"`

	// Nat IP address range. Field introduced in 18.2.3.
	NatIPRange *IPAddrRange `json:"nat_ip_range,omitempty"`
}

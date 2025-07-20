// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NatAddrInfo nat addr info
// swagger:model NatAddrInfo
type NatAddrInfo struct {

	// Nat IP address. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NatIP *IPAddr `json:"nat_ip,omitempty"`

	// Nat IP address range. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NatIPRange *IPAddrRange `json:"nat_ip_range,omitempty"`
}

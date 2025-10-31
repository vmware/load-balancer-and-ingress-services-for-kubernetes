// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAddrRange Ip addr range
// swagger:model IpAddrRange
type IPAddrRange struct {

	// Starting IP address of the range. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Begin *IPAddr `json:"begin"`

	// Ending IP address of the range. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	End *IPAddr `json:"end"`
}

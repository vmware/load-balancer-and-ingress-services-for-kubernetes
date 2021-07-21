// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAddr Ip addr
// swagger:model IpAddr
type IPAddr struct {

	// IP address.
	// Required: true
	Addr *string `json:"addr"`

	//  Enum options - V4, DNS, V6.
	// Required: true
	Type *string `json:"type"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugIPAddr debug Ip addr
// swagger:model DebugIpAddr
type DebugIPAddr struct {

	// Placeholder for description of property addrs of obj type DebugIpAddr field type str  type object
	Addrs []*IPAddr `json:"addrs,omitempty"`

	// Placeholder for description of property prefixes of obj type DebugIpAddr field type str  type object
	Prefixes []*IPAddrPrefix `json:"prefixes,omitempty"`

	// Placeholder for description of property ranges of obj type DebugIpAddr field type str  type object
	Ranges []*IPAddrRange `json:"ranges,omitempty"`
}

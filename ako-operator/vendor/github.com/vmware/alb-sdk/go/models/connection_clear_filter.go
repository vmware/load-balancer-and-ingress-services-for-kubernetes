// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConnectionClearFilter connection clear filter
// swagger:model ConnectionClearFilter
type ConnectionClearFilter struct {

	// IP address in dotted decimal notation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAddr *string `json:"ip_addr,omitempty"`

	// Port number. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *uint32 `json:"port,omitempty"`
}

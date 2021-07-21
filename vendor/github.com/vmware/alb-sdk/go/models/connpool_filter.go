// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConnpoolFilter connpool filter
// swagger:model ConnpoolFilter
type ConnpoolFilter struct {

	// Backend or SE IP address.
	IPAddr *string `json:"ip_addr,omitempty"`

	// Backend or SE IP address mask.
	IPMask *string `json:"ip_mask,omitempty"`

	// Backend or SE port.
	Port *int32 `json:"port,omitempty"`

	// cache type. Enum options - CP_ALL, CP_FREE, CP_BIND, CP_CACHED.
	Type *string `json:"type,omitempty"`
}

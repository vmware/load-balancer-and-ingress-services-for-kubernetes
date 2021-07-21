// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAllocInfo Ip alloc info
// swagger:model IpAllocInfo
type IPAllocInfo struct {

	// Placeholder for description of property ip of obj type IpAllocInfo field type str  type object
	// Required: true
	IP *IPAddr `json:"ip"`

	// mac of IpAllocInfo.
	// Required: true
	Mac *string `json:"mac"`

	// Unique object identifier of se.
	// Required: true
	SeUUID *string `json:"se_uuid"`
}

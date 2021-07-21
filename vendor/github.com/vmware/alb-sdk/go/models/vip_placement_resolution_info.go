// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipPlacementResolutionInfo vip placement resolution info
// swagger:model VipPlacementResolutionInfo
type VipPlacementResolutionInfo struct {

	// Placeholder for description of property ip of obj type VipPlacementResolutionInfo field type str  type object
	IP *IPAddr `json:"ip,omitempty"`

	// Placeholder for description of property networks of obj type VipPlacementResolutionInfo field type str  type object
	Networks []*DiscoveredNetwork `json:"networks,omitempty"`

	// Unique object identifier of pool.
	PoolUUID *string `json:"pool_uuid,omitempty"`
}

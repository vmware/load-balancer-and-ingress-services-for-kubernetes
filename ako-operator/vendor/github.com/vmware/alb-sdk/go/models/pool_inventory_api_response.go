// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolInventoryAPIResponse pool inventory Api response
// swagger:model PoolInventoryApiResponse
type PoolInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*PoolInventory `json:"results,omitempty"`
}

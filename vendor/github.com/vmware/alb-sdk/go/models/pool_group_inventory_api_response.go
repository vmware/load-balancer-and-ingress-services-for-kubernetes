// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolGroupInventoryAPIResponse pool group inventory Api response
// swagger:model PoolGroupInventoryApiResponse
type PoolGroupInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*PoolGroupInventory `json:"results,omitempty"`
}

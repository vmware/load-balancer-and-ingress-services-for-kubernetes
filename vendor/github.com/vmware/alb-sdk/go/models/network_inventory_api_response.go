// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkInventoryAPIResponse network inventory Api response
// swagger:model NetworkInventoryApiResponse
type NetworkInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NetworkInventory `json:"results,omitempty"`
}

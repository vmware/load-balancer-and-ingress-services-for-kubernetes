// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsvipInventoryAPIResponse vsvip inventory Api response
// swagger:model VsvipInventoryApiResponse
type VsvipInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VsvipInventory `json:"results,omitempty"`
}

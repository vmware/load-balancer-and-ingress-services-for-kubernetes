// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudInventoryAPIResponse cloud inventory Api response
// swagger:model CloudInventoryApiResponse
type CloudInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*CloudInventory `json:"results,omitempty"`
}

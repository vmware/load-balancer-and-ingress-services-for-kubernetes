// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsInventoryAPIResponse vs inventory Api response
// swagger:model VsInventoryApiResponse
type VsInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VsInventory `json:"results,omitempty"`
}

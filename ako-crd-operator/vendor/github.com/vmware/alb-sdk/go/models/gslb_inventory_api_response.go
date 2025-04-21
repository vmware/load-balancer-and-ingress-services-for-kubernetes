// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbInventoryAPIResponse gslb inventory Api response
// swagger:model GslbInventoryApiResponse
type GslbInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*GslbInventory `json:"results,omitempty"`
}

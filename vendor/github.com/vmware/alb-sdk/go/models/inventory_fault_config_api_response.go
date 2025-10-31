// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InventoryFaultConfigAPIResponse inventory fault config Api response
// swagger:model InventoryFaultConfigApiResponse
type InventoryFaultConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*InventoryFaultConfig `json:"results,omitempty"`
}

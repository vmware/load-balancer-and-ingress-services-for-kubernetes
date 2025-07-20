// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineGroupInventoryAPIResponse service engine group inventory Api response
// swagger:model ServiceEngineGroupInventoryApiResponse
type ServiceEngineGroupInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServiceEngineGroupInventory `json:"results,omitempty"`
}

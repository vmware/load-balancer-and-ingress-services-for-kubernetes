// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineInventoryAPIResponse service engine inventory Api response
// swagger:model ServiceEngineInventoryApiResponse
type ServiceEngineInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServiceEngineInventory `json:"results,omitempty"`
}

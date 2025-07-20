// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FederationCheckpointInventoryAPIResponse federation checkpoint inventory Api response
// swagger:model FederationCheckpointInventoryApiResponse
type FederationCheckpointInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*FederationCheckpointInventory `json:"results,omitempty"`
}

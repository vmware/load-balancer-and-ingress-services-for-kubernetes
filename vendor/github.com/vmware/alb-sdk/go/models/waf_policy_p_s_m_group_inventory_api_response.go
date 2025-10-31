// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicyPSMGroupInventoryAPIResponse waf policy p s m group inventory Api response
// swagger:model WafPolicyPSMGroupInventoryApiResponse
type WafPolicyPSMGroupInventoryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*WafPolicyPSMGroupInventory `json:"results,omitempty"`
}

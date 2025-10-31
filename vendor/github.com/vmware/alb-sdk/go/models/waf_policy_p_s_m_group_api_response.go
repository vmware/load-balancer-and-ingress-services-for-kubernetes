// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicyPSMGroupAPIResponse waf policy p s m group Api response
// swagger:model WafPolicyPSMGroupApiResponse
type WafPolicyPSMGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*WafPolicyPSMGroup `json:"results,omitempty"`
}

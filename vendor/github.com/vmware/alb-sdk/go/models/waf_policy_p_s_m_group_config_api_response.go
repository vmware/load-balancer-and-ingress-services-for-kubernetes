// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicyPSMGroupConfigAPIResponse waf policy p s m group config Api response
// swagger:model WafPolicyPSMGroupConfigApiResponse
type WafPolicyPSMGroupConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*WafPolicyPSMGroupConfig `json:"results,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafCRSAPIResponse waf c r s Api response
// swagger:model WafCRSApiResponse
type WafCRSAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*WafCRS `json:"results,omitempty"`
}

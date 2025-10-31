// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPPolicySetAPIResponse HTTP policy set Api response
// swagger:model HTTPPolicySetApiResponse
type HTTPPolicySetAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*HTTPPolicySet `json:"results,omitempty"`
}

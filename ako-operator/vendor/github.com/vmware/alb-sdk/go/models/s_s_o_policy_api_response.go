// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSOPolicyAPIResponse s s o policy Api response
// swagger:model SSOPolicyApiResponse
type SSOPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SSOPolicy `json:"results,omitempty"`
}

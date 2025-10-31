// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecurityPolicyAPIResponse security policy Api response
// swagger:model SecurityPolicyApiResponse
type SecurityPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SecurityPolicy `json:"results,omitempty"`
}

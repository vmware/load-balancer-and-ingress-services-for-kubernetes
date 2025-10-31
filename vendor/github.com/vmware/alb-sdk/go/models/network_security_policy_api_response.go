// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkSecurityPolicyAPIResponse network security policy Api response
// swagger:model NetworkSecurityPolicyApiResponse
type NetworkSecurityPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NetworkSecurityPolicy `json:"results,omitempty"`
}

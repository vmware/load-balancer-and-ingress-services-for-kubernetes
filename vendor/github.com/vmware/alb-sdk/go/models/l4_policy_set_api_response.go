// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L4PolicySetAPIResponse l4 policy set Api response
// swagger:model L4PolicySetApiResponse
type L4PolicySetAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*L4PolicySet `json:"results,omitempty"`
}

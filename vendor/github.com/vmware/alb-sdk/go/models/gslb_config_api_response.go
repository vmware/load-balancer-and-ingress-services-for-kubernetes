// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbConfigAPIResponse gslb config Api response
// swagger:model GslbConfigApiResponse
type GslbConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*GslbConfig `json:"results,omitempty"`
}

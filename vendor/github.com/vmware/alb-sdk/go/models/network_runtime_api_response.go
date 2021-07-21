// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkRuntimeAPIResponse network runtime Api response
// swagger:model NetworkRuntimeApiResponse
type NetworkRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NetworkRuntime `json:"results,omitempty"`
}

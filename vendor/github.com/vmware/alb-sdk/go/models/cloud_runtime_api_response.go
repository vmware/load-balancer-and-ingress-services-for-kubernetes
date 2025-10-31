// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudRuntimeAPIResponse cloud runtime Api response
// swagger:model CloudRuntimeApiResponse
type CloudRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*CloudRuntime `json:"results,omitempty"`
}

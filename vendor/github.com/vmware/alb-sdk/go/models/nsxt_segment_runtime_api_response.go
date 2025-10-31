// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtSegmentRuntimeAPIResponse nsxt segment runtime Api response
// swagger:model NsxtSegmentRuntimeApiResponse
type NsxtSegmentRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NsxtSegmentRuntime `json:"results,omitempty"`
}

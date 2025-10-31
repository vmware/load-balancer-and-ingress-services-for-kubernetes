// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SystemLimitsAPIResponse system limits Api response
// swagger:model SystemLimitsApiResponse
type SystemLimitsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SystemLimits `json:"results,omitempty"`
}

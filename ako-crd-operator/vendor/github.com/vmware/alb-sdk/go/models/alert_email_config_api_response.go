// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertEmailConfigAPIResponse alert email config Api response
// swagger:model AlertEmailConfigApiResponse
type AlertEmailConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AlertEmailConfig `json:"results,omitempty"`
}

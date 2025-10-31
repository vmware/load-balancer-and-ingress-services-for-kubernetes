// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugControllerAPIResponse debug controller Api response
// swagger:model DebugControllerApiResponse
type DebugControllerAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*DebugController `json:"results,omitempty"`
}

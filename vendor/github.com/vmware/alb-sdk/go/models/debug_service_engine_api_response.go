// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugServiceEngineAPIResponse debug service engine Api response
// swagger:model DebugServiceEngineApiResponse
type DebugServiceEngineAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*DebugServiceEngine `json:"results,omitempty"`
}

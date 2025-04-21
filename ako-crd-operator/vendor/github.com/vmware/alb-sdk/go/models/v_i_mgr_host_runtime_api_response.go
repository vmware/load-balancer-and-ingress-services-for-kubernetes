// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrHostRuntimeAPIResponse v i mgr host runtime Api response
// swagger:model VIMgrHostRuntimeApiResponse
type VIMgrHostRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIMgrHostRuntime `json:"results,omitempty"`
}

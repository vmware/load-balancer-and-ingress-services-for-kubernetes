// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrClusterRuntimeAPIResponse v i mgr cluster runtime Api response
// swagger:model VIMgrClusterRuntimeApiResponse
type VIMgrClusterRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIMgrClusterRuntime `json:"results,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrSEVMRuntimeAPIResponse v i mgr s e VM runtime Api response
// swagger:model VIMgrSEVMRuntimeApiResponse
type VIMgrSEVMRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIMgrSEVMRuntime `json:"results,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrVMRuntimeAPIResponse v i mgr VM runtime Api response
// swagger:model VIMgrVMRuntimeApiResponse
type VIMgrVMRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIMgrVMRuntime `json:"results,omitempty"`
}

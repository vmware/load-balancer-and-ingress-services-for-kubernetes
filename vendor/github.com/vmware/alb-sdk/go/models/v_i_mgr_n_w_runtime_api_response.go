// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrNWRuntimeAPIResponse v i mgr n w runtime Api response
// swagger:model VIMgrNWRuntimeApiResponse
type VIMgrNWRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIMgrNWRuntime `json:"results,omitempty"`
}

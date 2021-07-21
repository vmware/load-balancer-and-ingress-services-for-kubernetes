// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIMgrVcenterRuntimeAPIResponse v i mgr vcenter runtime Api response
// swagger:model VIMgrVcenterRuntimeApiResponse
type VIMgrVcenterRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIMgrVcenterRuntime `json:"results,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// APICLifsRuntimeAPIResponse API c lifs runtime Api response
// swagger:model APICLifsRuntimeApiResponse
type APICLifsRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*APICLifsRuntime `json:"results,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FileObjectDetailsAPIResponse file object details Api response
// swagger:model FileObjectDetailsApiResponse
type FileObjectDetailsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*FileObjectDetails `json:"results,omitempty"`
}

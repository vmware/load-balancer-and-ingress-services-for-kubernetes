// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FileObjectAPIResponse file object Api response
// swagger:model FileObjectApiResponse
type FileObjectAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*FileObject `json:"results,omitempty"`
}

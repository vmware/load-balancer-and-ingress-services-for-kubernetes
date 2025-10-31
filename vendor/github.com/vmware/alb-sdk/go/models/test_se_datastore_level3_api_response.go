// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TestSeDatastoreLevel3APIResponse test se datastore level3 Api response
// swagger:model TestSeDatastoreLevel3ApiResponse
type TestSeDatastoreLevel3APIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TestSeDatastoreLevel3 `json:"results,omitempty"`
}

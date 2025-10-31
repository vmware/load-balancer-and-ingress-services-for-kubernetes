// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TestSeDatastoreLevel1APIResponse test se datastore level1 Api response
// swagger:model TestSeDatastoreLevel1ApiResponse
type TestSeDatastoreLevel1APIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TestSeDatastoreLevel1 `json:"results,omitempty"`
}

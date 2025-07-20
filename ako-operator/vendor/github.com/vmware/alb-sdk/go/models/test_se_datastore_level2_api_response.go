// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TestSeDatastoreLevel2APIResponse test se datastore level2 Api response
// swagger:model TestSeDatastoreLevel2ApiResponse
type TestSeDatastoreLevel2APIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TestSeDatastoreLevel2 `json:"results,omitempty"`
}

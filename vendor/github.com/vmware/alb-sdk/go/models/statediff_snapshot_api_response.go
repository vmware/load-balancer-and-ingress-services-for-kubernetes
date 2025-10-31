// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// StatediffSnapshotAPIResponse statediff snapshot Api response
// swagger:model StatediffSnapshotApiResponse
type StatediffSnapshotAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*StatediffSnapshot `json:"results,omitempty"`
}

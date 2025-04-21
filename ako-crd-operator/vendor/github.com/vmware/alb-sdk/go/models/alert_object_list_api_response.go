// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertObjectListAPIResponse alert object list Api response
// swagger:model AlertObjectListApiResponse
type AlertObjectListAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AlertObjectList `json:"results,omitempty"`
}

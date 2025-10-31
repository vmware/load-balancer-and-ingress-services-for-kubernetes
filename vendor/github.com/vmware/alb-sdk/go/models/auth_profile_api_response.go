// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthProfileAPIResponse auth profile Api response
// swagger:model AuthProfileApiResponse
type AuthProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AuthProfile `json:"results,omitempty"`
}

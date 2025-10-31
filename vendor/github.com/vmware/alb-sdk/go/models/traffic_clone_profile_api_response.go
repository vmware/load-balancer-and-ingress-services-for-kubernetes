// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TrafficCloneProfileAPIResponse traffic clone profile Api response
// swagger:model TrafficCloneProfileApiResponse
type TrafficCloneProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*TrafficCloneProfile `json:"results,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VrfContextAPIResponse vrf context Api response
// swagger:model VrfContextApiResponse
type VrfContextAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VrfContext `json:"results,omitempty"`
}

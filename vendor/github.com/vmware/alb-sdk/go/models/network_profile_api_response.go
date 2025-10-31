// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkProfileAPIResponse network profile Api response
// swagger:model NetworkProfileApiResponse
type NetworkProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*NetworkProfile `json:"results,omitempty"`
}

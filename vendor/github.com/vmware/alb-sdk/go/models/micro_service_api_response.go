// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MicroServiceAPIResponse micro service Api response
// swagger:model MicroServiceApiResponse
type MicroServiceAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*MicroService `json:"results,omitempty"`
}

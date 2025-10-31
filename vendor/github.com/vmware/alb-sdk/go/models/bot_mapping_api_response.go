// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotMappingAPIResponse bot mapping Api response
// swagger:model BotMappingApiResponse
type BotMappingAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BotMapping `json:"results,omitempty"`
}

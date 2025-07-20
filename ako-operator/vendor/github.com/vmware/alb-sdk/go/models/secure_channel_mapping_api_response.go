// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecureChannelMappingAPIResponse secure channel mapping Api response
// swagger:model SecureChannelMappingApiResponse
type SecureChannelMappingAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SecureChannelMapping `json:"results,omitempty"`
}

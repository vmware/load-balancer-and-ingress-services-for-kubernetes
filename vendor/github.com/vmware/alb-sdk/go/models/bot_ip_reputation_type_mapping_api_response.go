// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotIPReputationTypeMappingAPIResponse bot IP reputation type mapping Api response
// swagger:model BotIPReputationTypeMappingApiResponse
type BotIPReputationTypeMappingAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BotIPReputationTypeMapping `json:"results,omitempty"`
}

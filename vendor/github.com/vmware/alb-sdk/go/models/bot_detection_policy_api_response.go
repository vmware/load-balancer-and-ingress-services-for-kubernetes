// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotDetectionPolicyAPIResponse bot detection policy Api response
// swagger:model BotDetectionPolicyApiResponse
type BotDetectionPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BotDetectionPolicy `json:"results,omitempty"`
}

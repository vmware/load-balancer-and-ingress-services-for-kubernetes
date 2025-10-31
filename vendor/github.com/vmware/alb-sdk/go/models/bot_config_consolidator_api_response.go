// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotConfigConsolidatorAPIResponse bot config consolidator Api response
// swagger:model BotConfigConsolidatorApiResponse
type BotConfigConsolidatorAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BotConfigConsolidator `json:"results,omitempty"`
}

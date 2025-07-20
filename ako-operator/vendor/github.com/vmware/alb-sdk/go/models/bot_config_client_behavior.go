// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotConfigClientBehavior bot config client behavior
// swagger:model BotConfigClientBehavior
type BotConfigClientBehavior struct {

	// Minimum percentage of bad requests for the client behavior component to identify as a bot. Allowed values are 1-100. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BadRequestPercent *uint32 `json:"bad_request_percent,omitempty"`

	// Whether client behavior based Bot detection is enabled. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Minimum requests for the client behavior component to make a decision. Allowed values are 2-1000. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinimumRequests *uint64 `json:"minimum_requests,omitempty"`

	// Minimum requests with a referer header for the client behavior component to not identify as a bot. Setting this to zero means the component never identifies a client as bot based on missing referer headers. Allowed values are 0-100. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinimumRequestsWithReferer *uint64 `json:"minimum_requests_with_referer,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotConfigClientBehavior bot config client behavior
// swagger:model BotConfigClientBehavior
type BotConfigClientBehavior struct {

	// Minimum percentage of bad requests for the client behavior component to identify as a bot. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BadRequestPercent *uint32 `json:"bad_request_percent,omitempty"`

	// Whether client behavior based Bot detection is enabled. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Minimum requests for the client behavior component to make a decision. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinimumRequests *uint64 `json:"minimum_requests,omitempty"`

	// Minimum requests with a referer for the client behavior component to not identify as a bot. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinimumRequestsWithReferer *uint64 `json:"minimum_requests_with_referer,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotDetectionMatch bot detection match
// swagger:model BotDetectionMatch
type BotDetectionMatch struct {

	// Bot classification types. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Classifications []*BotClassification `json:"classifications,omitempty"`

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MatchOperation *string `json:"match_operation"`
}

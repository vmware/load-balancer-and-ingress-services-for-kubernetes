// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MicroServiceMatch micro service match
// swagger:model MicroServiceMatch
type MicroServiceMatch struct {

	// UUID of Micro Service group(s). It is a reference to an object of type MicroServiceGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	GroupRef *string `json:"group_ref"`

	// Criterion to use for Micro Service matching the HTTP request. Enum options - IS_IN, IS_NOT_IN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`
}

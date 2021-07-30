// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MicroServiceMatch micro service match
// swagger:model MicroServiceMatch
type MicroServiceMatch struct {

	// UUID of Micro Service group(s). It is a reference to an object of type MicroServiceGroup.
	// Required: true
	GroupRef *string `json:"group_ref"`

	// Criterion to use for Micro Service matching the HTTP request. Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`
}

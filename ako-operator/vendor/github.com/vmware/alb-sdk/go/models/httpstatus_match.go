// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpstatusMatch httpstatus match
// swagger:model HTTPStatusMatch
type HttpstatusMatch struct {

	// Criterion to use for matching the HTTP response status code(s). Enum options - IS_IN, IS_NOT_IN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// HTTP response status code range(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ranges []*HttpstatusRange `json:"ranges,omitempty"`

	// HTTP response status code(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StatusCodes []int64 `json:"status_codes,omitempty,omitempty"`
}

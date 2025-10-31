// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CookieMatch cookie match
// swagger:model CookieMatch
type CookieMatch struct {

	// Case sensitivity to use for the match. Enum options - SENSITIVE, INSENSITIVE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchCase *string `json:"match_case,omitempty"`

	// Criterion to use for matching the cookie in the HTTP request. Enum options - HDR_EXISTS, HDR_DOES_NOT_EXIST, HDR_BEGINS_WITH, HDR_DOES_NOT_BEGIN_WITH, HDR_CONTAINS, HDR_DOES_NOT_CONTAIN, HDR_ENDS_WITH, HDR_DOES_NOT_END_WITH, HDR_EQUALS, HDR_DOES_NOT_EQUAL. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Name of the cookie. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// String value in the cookie. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}

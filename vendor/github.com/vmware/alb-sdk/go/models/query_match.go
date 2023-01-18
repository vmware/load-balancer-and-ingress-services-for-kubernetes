// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// QueryMatch query match
// swagger:model QueryMatch
type QueryMatch struct {

	// Case sensitivity to use for the match. Enum options - SENSITIVE, INSENSITIVE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchCase *string `json:"match_case,omitempty"`

	// Criterion to use for matching the query in HTTP request URI. Enum options - QUERY_MATCH_CONTAINS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// String value(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the *string group(s). It is a reference to an object of type StringGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

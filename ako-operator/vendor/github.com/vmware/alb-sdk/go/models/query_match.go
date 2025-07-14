// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// QueryMatch query match
// swagger:model QueryMatch
type QueryMatch struct {

	// Case sensitivity to use for the match. Enum options - SENSITIVE, INSENSITIVE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchCase *string `json:"match_case,omitempty"`

	// Criterion to use for matching the query in HTTP request URI. Enum options - QUERY_MATCH_CONTAINS, QUERY_MATCH_DOES_NOT_CONTAIN, QUERY_MATCH_EXISTS, QUERY_MATCH_DOES_NOT_EXIST, QUERY_MATCH_BEGINS_WITH, QUERY_MATCH_DOES_NOT_BEGIN_WITH, QUERY_MATCH_ENDS_WITH, QUERY_MATCH_DOES_NOT_END_WITH, QUERY_MATCH_EQUALS, QUERY_MATCH_DOES_NOT_EQUAL, QUERY_MATCH_REGEX_MATCH, QUERY_MATCH_REGEX_DOES_NOT_MATCH. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Match against the decoded URI query. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchDecodedString *bool `json:"match_decoded_string,omitempty"`

	// String value(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the *string group(s). It is a reference to an object of type StringGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

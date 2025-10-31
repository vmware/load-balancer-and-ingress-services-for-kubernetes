// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PathMatch path match
// swagger:model PathMatch
type PathMatch struct {

	// Case sensitivity to use for the matching. Enum options - SENSITIVE, INSENSITIVE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchCase *string `json:"match_case,omitempty"`

	// Criterion to use for matching the path in the HTTP request URI. Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- BEGINS_WITH,DOES_NOT_BEGIN_WITH,CONTAINS,DOES_NOT_CONTAIN,ENDS_WITH,DOES_NOT_END_WITH,EQUALS,DOES_NOT_EQUAL), Basic edition(Allowed values- BEGINS_WITH,DOES_NOT_BEGIN_WITH,CONTAINS,DOES_NOT_CONTAIN,ENDS_WITH,DOES_NOT_END_WITH,EQUALS,DOES_NOT_EQUAL), Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Match against the decoded URI path. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchDecodedString *bool `json:"match_decoded_string,omitempty"`

	// String values. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the *string group(s). It is a reference to an object of type StringGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

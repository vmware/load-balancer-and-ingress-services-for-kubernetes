// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// StringMatch *string match
// swagger:model StringMatch
type StringMatch struct {

	// Criterion to use for *string matching the HTTP request. Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- BEGINS_WITH,DOES_NOT_BEGIN_WITH,CONTAINS,DOES_NOT_CONTAIN,ENDS_WITH,DOES_NOT_END_WITH,EQUALS,DOES_NOT_EQUAL), Basic edition(Allowed values- BEGINS_WITH,DOES_NOT_BEGIN_WITH,CONTAINS,DOES_NOT_CONTAIN,ENDS_WITH,DOES_NOT_END_WITH,EQUALS,DOES_NOT_EQUAL), Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// String value(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the *string group(s). It is a reference to an object of type StringGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

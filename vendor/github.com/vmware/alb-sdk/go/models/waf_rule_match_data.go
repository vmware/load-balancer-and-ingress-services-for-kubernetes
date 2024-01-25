// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafRuleMatchData waf rule match data
// swagger:model WafRuleMatchData
type WafRuleMatchData struct {

	// The match_element is an internal variable. It is not possible to add exclusions for this element. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsInternal *bool `json:"is_internal,omitempty"`

	// Field from a transaction that matches the rule, for instance if the request parameter is password=foobar, then match_element is ARGS password. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchElement *string `json:"match_element,omitempty"`

	// Value of the field from a transaction that matches the rule. For instance, if the request parameter is password=foo, then match_value is foo. The value can be truncated if it is too long. In this case, this field starts at the position where the actual match started inside the value, and that position is stored in match_value_offset. This is done to ensure the relevant part is shown. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchValue *string `json:"match_value,omitempty"`

	// The starting index of the first character of match_value field with respect to original match value. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchValueOffset uint64 `json:"match_value_offset,omitempty"`
}

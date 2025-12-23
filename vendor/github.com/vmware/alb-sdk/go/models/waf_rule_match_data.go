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

	// Value for a field from a transaction that matches the rule, for instance if the request parameter is password=foobar, then match_value is foobar. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchValue *string `json:"match_value,omitempty"`
}

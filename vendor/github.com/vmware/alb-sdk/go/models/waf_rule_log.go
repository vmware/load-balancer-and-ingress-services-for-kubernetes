// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafRuleLog waf rule log
// swagger:model WafRuleLog
type WafRuleLog struct {

	// Transaction data that matched the rule. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Matches []*WafRuleMatchData `json:"matches,omitempty"`

	// Rule's msg *string per ModSec language. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Msg *string `json:"msg,omitempty"`

	// The count of omitted match element logs in the current rule. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OmittedMatchElements uint32 `json:"omitted_match_elements,omitempty"`

	// Phase in which transaction matched the Rule - for instance, Request Header Phase. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Phase *string `json:"phase,omitempty"`

	// Rule Group for the matching rule. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RuleGroup *string `json:"rule_group,omitempty"`

	// ID of the matching rule per ModSec language. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RuleID uint64 `json:"rule_id,omitempty"`

	// Name of the rule. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RuleName *string `json:"rule_name,omitempty"`

	// Rule's tags per ModSec language. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tags []string `json:"tags,omitempty"`
}

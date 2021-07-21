// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafRuleGroupOverrides waf rule group overrides
// swagger:model WafRuleGroupOverrides
type WafRuleGroupOverrides struct {

	// Override the enable flag for this group. Field introduced in 20.1.6.
	Enable *bool `json:"enable,omitempty"`

	// Replace the exclude list for this group. Field introduced in 20.1.6. Maximum of 64 items allowed.
	ExcludeList []*WafExcludeListEntry `json:"exclude_list,omitempty"`

	// Override the waf mode for this group.. Enum options - WAF_MODE_DETECTION_ONLY, WAF_MODE_ENFORCEMENT. Field introduced in 20.1.6.
	Mode *string `json:"mode,omitempty"`

	// The name of the group where attributes or rules are overridden. Field introduced in 20.1.6.
	// Required: true
	Name *string `json:"name"`

	// Rule specific overrides. Field introduced in 20.1.6. Maximum of 1024 items allowed.
	RuleOverrides []*WafRuleOverrides `json:"rule_overrides,omitempty"`
}

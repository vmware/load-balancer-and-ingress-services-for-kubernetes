// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafRuleOverrides waf rule overrides
// swagger:model WafRuleOverrides
type WafRuleOverrides struct {

	// Override the enable flag for this rule. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`

	// Replace the exclude list for this rule. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ExcludeList []*WafExcludeListEntry `json:"exclude_list,omitempty"`

	// Override the waf mode for this rule. Enum options - WAF_MODE_DETECTION_ONLY, WAF_MODE_ENFORCEMENT. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Mode *string `json:"mode,omitempty"`

	// The rule_id of the rule where attributes are overridden. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	RuleID *string `json:"rule_id"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPSMRule waf p s m rule
// swagger:model WafPSMRule
type WafPSMRule struct {

	// Free-text comment about this rule. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable or disable this rule. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`

	// Rule index, this is used to determine the order of the rules. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	// The field match_value_pattern regular expression is case sensitive. Enum options - SENSITIVE, INSENSITIVE. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchCase *string `json:"match_case,omitempty"`

	// The match elements, for example ARGS id or ARGS|!ARGS password. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchElements []*WafPSMMatchElement `json:"match_elements,omitempty"`

	// The maximum allowed length of the match_value. If this is not set, the length will not be checked. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchValueMaxLength *uint32 `json:"match_value_max_length,omitempty"`

	// A regular expression which describes the expected value. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchValuePattern *string `json:"match_value_pattern,omitempty"`

	// If match_value_string_group_uuid and match_value_string_group_key are set, the referenced regular expression is used as match_value_pattern. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchValueStringGroupKey *string `json:"match_value_string_group_key,omitempty"`

	// The UUID of a *string group containing key used in match_value_string_group_key. It is a reference to an object of type StringGroup. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchValueStringGroupRef *string `json:"match_value_string_group_ref,omitempty"`

	// WAF Rule mode. This can be detection or enforcement. If this is not set, the Policy mode is used. This only takes effect if the policy allows delegation. Enum options - WAF_MODE_DETECTION_ONLY, WAF_MODE_ENFORCEMENT. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mode *string `json:"mode,omitempty"`

	// Name of the rule. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// WAF Ruleset paranoia mode. This is used to select Rules based on the paranoia-level. Enum options - WAF_PARANOIA_LEVEL_LOW, WAF_PARANOIA_LEVEL_MEDIUM, WAF_PARANOIA_LEVEL_HIGH, WAF_PARANOIA_LEVEL_EXTREME. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParanoiaLevel *string `json:"paranoia_level,omitempty"`

	// Id field which is used for log and metric generation. This id must be unique for all rules in this group. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RuleID *string `json:"rule_id"`
}

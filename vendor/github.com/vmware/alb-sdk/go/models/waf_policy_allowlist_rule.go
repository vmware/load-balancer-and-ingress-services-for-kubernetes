// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicyAllowlistRule waf policy allowlist rule
// swagger:model WafPolicyAllowlistRule
type WafPolicyAllowlistRule struct {

	// Actions to be performed upon successful matching. Enum options - WAF_POLICY_ALLOWLIST_ACTION_BYPASS, WAF_POLICY_ALLOWLIST_ACTION_DETECTION_MODE, WAF_POLICY_ALLOWLIST_ACTION_CONTINUE. Field introduced in 20.1.3. Minimum of 1 items required. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Actions []string `json:"actions,omitempty"`

	// Description of this rule. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable or deactivate the rule. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`

	// Rules are processed in order of this index field. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	// Match criteria describing requests to which this rule should be applied. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Match *MatchTarget `json:"match"`

	// A name describing the rule in a short form. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Percentage of traffic that is sampled. Allowed values are 0-100. Field introduced in 20.1.3. Unit is PERCENT. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SamplingPercent *uint32 `json:"sampling_percent,omitempty"`
}

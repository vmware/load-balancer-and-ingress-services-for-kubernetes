package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyAllowlistRule waf policy allowlist rule
// swagger:model WafPolicyAllowlistRule
type WafPolicyAllowlistRule struct {

	// Actions to be performed upon successful matching. Enum options - WAF_POLICY_ALLOWLIST_ACTION_BYPASS, WAF_POLICY_ALLOWLIST_ACTION_DETECTION_MODE, WAF_POLICY_ALLOWLIST_ACTION_CONTINUE. Field introduced in 20.1.3. Minimum of 1 items required. Maximum of 1 items allowed.
	Actions []string `json:"actions,omitempty"`

	// Description of this rule. Field introduced in 20.1.3.
	Description *string `json:"description,omitempty"`

	// Enable or deactivate the rule. Field introduced in 20.1.3.
	Enable *bool `json:"enable,omitempty"`

	// Rules are processed in order of this index field. Field introduced in 20.1.3.
	// Required: true
	Index *int32 `json:"index"`

	// Match criteria describing requests to which this rule should be applied. Field introduced in 20.1.3.
	// Required: true
	Match *MatchTarget `json:"match"`

	// A name describing the rule in a short form. Field introduced in 20.1.3.
	// Required: true
	Name *string `json:"name"`

	// Percentage of traffic that is sampled. Allowed values are 0-100. Field introduced in 20.1.3. Unit is PERCENT.
	SamplingPercent *int32 `json:"sampling_percent,omitempty"`
}

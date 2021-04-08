package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyWhitelistRule waf policy whitelist rule
// swagger:model WafPolicyWhitelistRule
type WafPolicyWhitelistRule struct {

	// Actions to be performed upon successful matching. Enum options - WAF_POLICY_WHITELIST_ACTION_ALLOW, WAF_POLICY_WHITELIST_ACTION_DETECTION_MODE, WAF_POLICY_WHITELIST_ACTION_CONTINUE. Field deprecated in 20.1.3. Field introduced in 18.2.3. Minimum of 1 items required. Maximum of 1 items allowed.
	Actions []string `json:"actions,omitempty"`

	// Description of this rule. Field deprecated in 20.1.3. Field introduced in 18.2.3.
	Description *string `json:"description,omitempty"`

	// Enable or disable the rule. Field deprecated in 20.1.3. Field introduced in 18.2.3.
	Enable *bool `json:"enable,omitempty"`

	// Rules are executed in order of this index field. Field deprecated in 20.1.3. Field introduced in 18.2.3.
	// Required: true
	Index *int32 `json:"index"`

	// Match criteria describing requests to which this rule should be applied. Field deprecated in 20.1.3. Field introduced in 18.2.3.
	// Required: true
	Match *MatchTarget `json:"match"`

	// A name describing the rule in a short form. Field deprecated in 20.1.3. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`

	// Percentage of traffic that is sampled. Allowed values are 0-100. Field deprecated in 20.1.3. Field introduced in 20.1.1. Unit is PERCENT.
	SamplingPercent *int32 `json:"sampling_percent,omitempty"`
}

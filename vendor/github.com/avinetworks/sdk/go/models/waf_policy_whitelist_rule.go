package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyWhitelistRule waf policy whitelist rule
// swagger:model WafPolicyWhitelistRule
type WafPolicyWhitelistRule struct {

	// Actions to be performed upon successful matching. Enum options - WAF_POLICY_WHITELIST_ACTION_ALLOW, WAF_POLICY_WHITELIST_ACTION_DETECTION_MODE, WAF_POLICY_WHITELIST_ACTION_CONTINUE. Field introduced in 18.2.3.
	Actions []string `json:"actions,omitempty"`

	// Description of this rule. Field introduced in 18.2.3.
	Description *string `json:"description,omitempty"`

	// Enable or disable the rule. Field introduced in 18.2.3.
	Enable *bool `json:"enable,omitempty"`

	// Rules are executed in order of this index field. Field introduced in 18.2.3.
	// Required: true
	Index *int32 `json:"index"`

	// Match criteria describing requests to which this rule should be applied. Field introduced in 18.2.3.
	// Required: true
	Match *MatchTarget `json:"match"`

	// A name describing the rule in a short form. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafRuleLog waf rule log
// swagger:model WafRuleLog
type WafRuleLog struct {

	// Transaction data that matched the rule. Field introduced in 17.2.1.
	Matches []*WafRuleMatchData `json:"matches,omitempty"`

	// Rule's msg *string per ModSec language. Field introduced in 17.2.1.
	Msg *string `json:"msg,omitempty"`

	// Phase in which transaction matched the Rule - for instance, Request Header Phase. Field introduced in 17.2.1.
	Phase *string `json:"phase,omitempty"`

	// Rule Group for the matching rule. Field introduced in 17.2.1.
	RuleGroup *string `json:"rule_group,omitempty"`

	// ID of the matching rule per ModSec language. Field introduced in 17.2.1.
	RuleID *int64 `json:"rule_id,omitempty"`

	// Name of the rule. Field introduced in 17.2.3.
	RuleName *string `json:"rule_name,omitempty"`

	// Rule's tags per ModSec language. Field introduced in 17.2.1.
	Tags []string `json:"tags,omitempty"`
}

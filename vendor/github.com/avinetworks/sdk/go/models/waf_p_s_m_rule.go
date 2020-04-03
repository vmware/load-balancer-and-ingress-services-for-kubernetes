package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPSMRule waf p s m rule
// swagger:model WafPSMRule
type WafPSMRule struct {

	// Free-text comment about this rule. Field introduced in 18.2.3.
	Description *string `json:"description,omitempty"`

	// Enable or disable this rule. Field introduced in 18.2.3.
	Enable *bool `json:"enable,omitempty"`

	// Rule index, this is used to determine the order of the rules. Field introduced in 18.2.3.
	// Required: true
	Index *int32 `json:"index"`

	// The field match_value_pattern regular expression is case sensitive. Enum options - SENSITIVE, INSENSITIVE. Field introduced in 18.2.3.
	MatchCase *string `json:"match_case,omitempty"`

	// The match elements, for example ARGS id or ARGS|!ARGS password. Field introduced in 18.2.3.
	MatchElements []*WafPSMMatchElement `json:"match_elements,omitempty"`

	// The maximum allowed length of the match_value. If this is not set, the length will not be checked. Field introduced in 18.2.3.
	MatchValueMaxLength *int32 `json:"match_value_max_length,omitempty"`

	// A regular expression which describes the expected value. Field introduced in 18.2.3.
	// Required: true
	MatchValuePattern *string `json:"match_value_pattern"`

	// WAF Rule mode. This can be detection or enforcement. If this is not set, the Policy mode is used. This only takes effect if the policy allows delegation. Enum options - WAF_MODE_DETECTION_ONLY, WAF_MODE_ENFORCEMENT. Field introduced in 18.2.3.
	Mode *string `json:"mode,omitempty"`

	// Name of the rule. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`

	// WAF Ruleset paranoia mode. This is used to select Rules based on the paranoia-level. Enum options - WAF_PARANOIA_LEVEL_LOW, WAF_PARANOIA_LEVEL_MEDIUM, WAF_PARANOIA_LEVEL_HIGH, WAF_PARANOIA_LEVEL_EXTREME. Field introduced in 18.2.3.
	ParanoiaLevel *string `json:"paranoia_level,omitempty"`

	// Id field which is used for log and metric generation. This id must be unique for all rules in this group. Field introduced in 18.2.3.
	// Required: true
	RuleID *string `json:"rule_id"`
}

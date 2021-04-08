package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthnRuleMatch authn rule match
// swagger:model AuthnRuleMatch
type AuthnRuleMatch struct {

	// Name of the executed Authentication rule Action. Enum options - SKIP_AUTHENTICATION, USE_DEFAULT_AUTHENTICATION. Field introduced in 20.1.3.
	RuleAction *string `json:"rule_action,omitempty"`

	// Name of the matched Authentication rule. Field introduced in 20.1.3.
	RuleName *string `json:"rule_name,omitempty"`
}

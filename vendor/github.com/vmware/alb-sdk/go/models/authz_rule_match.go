package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthzRuleMatch authz rule match
// swagger:model AuthzRuleMatch
type AuthzRuleMatch struct {

	// Name of the executed Authorization rule Action. Enum options - ALLOW_ACCESS, CLOSE_CONNECTION, HTTP_LOCAL_RESPONSE. Field introduced in 20.1.3.
	RuleAction *string `json:"rule_action,omitempty"`

	// Name of the matched Authorization rule. Field introduced in 20.1.3.
	RuleName *string `json:"rule_name,omitempty"`
}

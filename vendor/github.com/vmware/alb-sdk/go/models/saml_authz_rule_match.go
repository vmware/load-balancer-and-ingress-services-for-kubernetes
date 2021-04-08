package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SamlAuthzRuleMatch saml authz rule match
// swagger:model SamlAuthzRuleMatch
type SamlAuthzRuleMatch struct {

	// Name of the executed SAML Authorization rule Action. Enum options - ALLOW_ACCESS, CLOSE_CONNECTION, HTTP_LOCAL_RESPONSE. Field introduced in 20.1.1.
	SamlAuthzMatchedRuleAction *string `json:"saml_authz_matched_rule_action,omitempty"`

	// Name of the matched SAML Authorization rule. Field introduced in 20.1.1.
	SamlAuthzMatchedRuleName *string `json:"saml_authz_matched_rule_name,omitempty"`
}

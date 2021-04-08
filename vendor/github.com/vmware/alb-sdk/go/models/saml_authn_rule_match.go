package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SamlAuthnRuleMatch saml authn rule match
// swagger:model SamlAuthnRuleMatch
type SamlAuthnRuleMatch struct {

	// Name of the executed SAML Authentication rule Action. Enum options - SKIP_AUTHENTICATION, USE_DEFAULT_AUTHENTICATION. Field introduced in 20.1.1.
	SamlAuthnMatchedRuleAction *string `json:"saml_authn_matched_rule_action,omitempty"`

	// Name of the matched SAML Authentication rule. Field introduced in 20.1.1.
	SamlAuthnMatchedRuleName *string `json:"saml_authn_matched_rule_name,omitempty"`
}

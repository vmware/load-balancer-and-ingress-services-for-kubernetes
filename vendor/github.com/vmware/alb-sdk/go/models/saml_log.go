package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SamlLog saml log
// swagger:model SamlLog
type SamlLog struct {

	// Set to True if SAML Authentication is used. Field introduced in 20.1.1.
	IsSamlAuthenticationUsed *bool `json:"is_saml_authentication_used,omitempty"`

	// SAML Attribute list. Field introduced in 20.1.1.
	SamlAttributeLists []*SamlAttribute `json:"saml_attribute_lists,omitempty"`

	// Saml Authentication Status. Enum options - SAML_AUTH_STATUS_UNAVAILABLE, SAML_AUTH_STATUS_UNAUTH_GET_REQUEST, SAML_AUTH_STATUS_UNAUTH_REQ_UNSUPPORTED_METHOD, SAML_AUTH_STATUS_AUTH_REQUEST_GENERATED, SAML_AUTH_STATUS_AUTH_RESPONSE_RECEIVED, SAML_AUTH_STATUS_AUTHENTICATED_REQUEST, SAML_AUTH_STATUS_AUTHORIZATION_FAILED. Field introduced in 20.1.1.
	SamlAuthStatus *string `json:"saml_auth_status,omitempty"`

	// SAML Authentication rule match. Field introduced in 20.1.1.
	SamlAuthnRuleMatch *SamlAuthnRuleMatch `json:"saml_authn_rule_match,omitempty"`

	// SAML Authorization rule match. Field introduced in 20.1.1.
	SamlAuthzRuleMatch *SamlAuthzRuleMatch `json:"saml_authz_rule_match,omitempty"`

	// Is set when SAML session cookie is expired. Field introduced in 20.1.1.
	SamlSessionCookieExpired *bool `json:"saml_session_cookie_expired,omitempty"`

	// SAML userid. Field introduced in 20.1.1.
	Userid *string `json:"userid,omitempty"`
}

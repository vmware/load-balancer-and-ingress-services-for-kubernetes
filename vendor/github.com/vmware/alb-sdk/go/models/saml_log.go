// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlLog saml log
// swagger:model SamlLog
type SamlLog struct {

	// Set to True if SAML Authentication is used. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsSamlAuthenticationUsed *bool `json:"is_saml_authentication_used,omitempty"`

	// SAML Attribute list. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAttributeLists []*SamlAttribute `json:"saml_attribute_lists,omitempty"`

	// Saml Authentication Status. Enum options - SAML_AUTH_STATUS_UNAVAILABLE, SAML_AUTH_STATUS_UNAUTH_GET_REQUEST, SAML_AUTH_STATUS_UNAUTH_REQ_UNSUPPORTED_METHOD, SAML_AUTH_STATUS_AUTH_REQUEST_GENERATED, SAML_AUTH_STATUS_AUTH_RESPONSE_RECEIVED, SAML_AUTH_STATUS_AUTHENTICATED_REQUEST, SAML_AUTH_STATUS_AUTHORIZATION_FAILED. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthStatus *string `json:"saml_auth_status,omitempty"`

	// SAML Authentication rule match. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthnRuleMatch *SamlAuthnRuleMatch `json:"saml_authn_rule_match,omitempty"`

	// SAML Authorization rule match. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthzRuleMatch *SamlAuthzRuleMatch `json:"saml_authz_rule_match,omitempty"`

	// Is set when SAML session cookie is expired. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlSessionCookieExpired *bool `json:"saml_session_cookie_expired,omitempty"`

	// SAML userid. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Userid *string `json:"userid,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OauthLog oauth log
// swagger:model OauthLog
type OauthLog struct {

	// Authentication policy rule match. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AuthnRuleMatch *AuthnRuleMatch `json:"authn_rule_match,omitempty"`

	// Authorization policy rule match. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AuthzRuleMatch *AuthzRuleMatch `json:"authz_rule_match,omitempty"`

	// OAuth SessionCookie expired. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IsSessionCookieExpired *bool `json:"is_session_cookie_expired,omitempty"`

	// Subrequest info related to fetching jwks keys from jwks uri endpoint. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	JwksSubrequest *OauthSubRequestLog `json:"jwks_subrequest,omitempty"`

	// OAuth state. Enum options - OAUTH_STATE_CLIENT_IDP_HANDSHAKE_REDIRECT, OAUTH_STATE_CLIENT_IDP_HANDSHAKE_FAIL, OAUTH_STATE_TOKEN_EXCHANGE_REQUEST, OAUTH_STATE_TOKEN_EXCHANGE_RESPONSE, OAUTH_STATE_TOKEN_INTROSPECTION_REQUEST, OAUTH_STATE_TOKEN_INTROSPECTION_RESPONSE, OAUTH_STATE_REFRESH_TOKEN_REQUEST, OAUTH_STATE_REFRESH_TOKEN_RESPONSE, OAUTH_STATE_JWKS_URI_REQUEST, OAUTH_STATE_JWKS_URI_RESPONSE, OAUTH_STATE_USERINFO_REQUEST, OAUTH_STATE_USERINFO_RESPONSE. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthState *string `json:"oauth_state,omitempty"`

	// OAuth request State to avoid CSRF atatcks. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`

	// Subrequest info related to the code exchange flow. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenExchangeSubrequest *OauthSubRequestLog `json:"token_exchange_subrequest,omitempty"`

	// Subrequest info related to Token Introspection. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenIntrospectionSubrequest *OauthSubRequestLog `json:"token_introspection_subrequest,omitempty"`

	// Subrequest info related to refresh access Token flow. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenRefreshSubrequest *OauthSubRequestLog `json:"token_refresh_subrequest,omitempty"`

	// Subrequest info related to fetching userinfo from userinfo endpoint. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UserinfoSubrequest *OauthSubRequestLog `json:"userinfo_subrequest,omitempty"`
}

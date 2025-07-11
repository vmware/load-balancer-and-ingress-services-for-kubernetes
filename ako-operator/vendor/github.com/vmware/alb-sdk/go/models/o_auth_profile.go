// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OAuthProfile o auth profile
// swagger:model OAuthProfile
type OAuthProfile struct {

	// URL of authorization server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	AuthorizationEndpoint *string `json:"authorization_endpoint"`

	// Logout URI of IDP server. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndSessionEndpoint *string `json:"end_session_endpoint,omitempty"`

	// Instance uuid of the csp service. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	InstanceID *string `json:"instance_id,omitempty"`

	// URL of token introspection server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IntrospectionEndpoint *string `json:"introspection_endpoint,omitempty"`

	// Uniquely identifiable name of the Token Issuer. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Issuer *string `json:"issuer,omitempty"`

	// Lifetime of the cached JWKS keys. Allowed values are 0-1440. Field introduced in 21.1.3. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	JwksTimeout *int32 `json:"jwks_timeout,omitempty"`

	// JWKS URL of the endpoint that hosts the public keys that can be used to verify any JWT issued by the authorization server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	JwksURI *string `json:"jwks_uri,omitempty"`

	// OAuth App Settings for controller authentication. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthControllerSettings *OAuthAppSettings `json:"oauth_controller_settings,omitempty"`

	// Type of OAuth profile which defines the usage type. Enum options - CLIENT_OAUTH, CONTROLLER_OAUTH. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthProfileType *string `json:"oauth_profile_type,omitempty"`

	// Type of OAuth Provider when using controller oauth as oauth profile type. Enum options - OAUTH_CSP. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthProvider *string `json:"oauth_provider,omitempty"`

	// Buffering size for the responses from the OAUTH enpoints. Allowed values are 0-32768000. Field introduced in 21.1.3. Unit is BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthRespBufferSz *int32 `json:"oauth_resp_buffer_sz,omitempty"`

	// Organization Id for OAuth. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OrgID *string `json:"org_id,omitempty"`

	// Pool object to interface with Authorization Server endpoints. It is a reference to an object of type Pool. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PoolRef *string `json:"pool_ref,omitempty"`

	// Redirect URI specified in the request to Authorization Server. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RedirectURI *string `json:"redirect_uri,omitempty"`

	// Uuid value of csp service. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceID *string `json:"service_id,omitempty"`

	// Name of the csp service. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceName *string `json:"service_name,omitempty"`

	// URL of token exchange server. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenEndpoint *string `json:"token_endpoint,omitempty"`

	// URL of the Userinfo Endpoint. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UserinfoEndpoint *string `json:"userinfo_endpoint,omitempty"`
}

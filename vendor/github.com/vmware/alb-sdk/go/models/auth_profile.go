// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthProfile auth profile
// swagger:model AuthProfile
type AuthProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// HTTP user authentication params. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTP *AuthProfileHTTPClientParams `json:"http,omitempty"`

	// JWTServerProfile to be used for authentication. It is a reference to an object of type JWTServerProfile. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	JwtProfileRef *string `json:"jwt_profile_ref,omitempty"`

	// LDAP server and directory settings. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ldap *LdapAuthSettings `json:"ldap,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Name of the Auth Profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// OAuth Profile - Common endpoint information. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthProfile *OAuthProfile `json:"oauth_profile,omitempty"`

	// SAML settings. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Saml *SamlSettings `json:"saml,omitempty"`

	// TACACS+ settings. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TacacsPlus *TacacsPlusAuthSettings `json:"tacacs_plus,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Type of the Auth Profile. Enum options - AUTH_PROFILE_LDAP, AUTH_PROFILE_TACACS_PLUS, AUTH_PROFILE_SAML, AUTH_PROFILE_PINGACCESS, AUTH_PROFILE_JWT, AUTH_PROFILE_OAUTH. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- AUTH_PROFILE_LDAP,AUTH_PROFILE_TACACS_PLUS,AUTH_PROFILE_SAML,AUTH_PROFILE_JWT,AUTH_PROFILE_OAUTH), Basic edition(Allowed values- AUTH_PROFILE_LDAP,AUTH_PROFILE_TACACS_PLUS,AUTH_PROFILE_SAML,AUTH_PROFILE_JWT,AUTH_PROFILE_OAUTH), Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the Auth Profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

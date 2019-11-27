package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthProfile auth profile
// swagger:model AuthProfile
type AuthProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// HTTP user authentication params.
	HTTP *AuthProfileHTTPClientParams `json:"http,omitempty"`

	// LDAP server and directory settings.
	Ldap *LdapAuthSettings `json:"ldap,omitempty"`

	// Name of the Auth Profile.
	// Required: true
	Name *string `json:"name"`

	// SAML settings. Field introduced in 17.2.3.
	Saml *SamlSettings `json:"saml,omitempty"`

	// TACACS+ settings.
	TacacsPlus *TacacsPlusAuthSettings `json:"tacacs_plus,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Type of the Auth Profile. Enum options - AUTH_PROFILE_LDAP, AUTH_PROFILE_TACACS_PLUS, AUTH_PROFILE_SAML.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the Auth Profile.
	UUID *string `json:"uuid,omitempty"`
}

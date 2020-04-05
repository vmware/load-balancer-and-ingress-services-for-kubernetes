package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthenticationPolicy authentication policy
// swagger:model AuthenticationPolicy
type AuthenticationPolicy struct {

	// Auth Profile to use for validating users. It is a reference to an object of type AuthProfile. Field deprecated in 18.2.3. Field introduced in 18.2.1.
	AuthProfileRef *string `json:"auth_profile_ref,omitempty"`

	// Add rules to apply auth profile to specific targets. Field introduced in 18.2.5.
	AuthnRules []*AuthenticationRule `json:"authn_rules,omitempty"`

	// HTTP cookie name for authenticated session. Field deprecated in 18.2.3. Field introduced in 18.2.1.
	CookieName *string `json:"cookie_name,omitempty"`

	// Cookie timeout in minutes. Allowed values are 1-1440. Field deprecated in 18.2.3. Field introduced in 18.2.1.
	CookieTimeout *int32 `json:"cookie_timeout,omitempty"`

	// Auth Profile to use for validating users. It is a reference to an object of type AuthProfile. Field introduced in 18.2.3.
	// Required: true
	DefaultAuthProfileRef *string `json:"default_auth_profile_ref"`

	// Globally unique entityID for this node. Entity ID on the IDP should match this. Field deprecated in 18.2.3. Field introduced in 18.2.1.
	EntityID *string `json:"entity_id,omitempty"`

	// Key to generate the cookie. Field deprecated in 18.2.3. Field introduced in 18.2.1.
	Key []*HTTPCookiePersistenceKey `json:"key,omitempty"`

	// Single Signon URL to be programmed on the IDP. Field deprecated in 18.2.3. Field introduced in 18.2.1.
	SingleSignonURL *string `json:"single_signon_url,omitempty"`

	// SAML SP metadata. Field deprecated in 18.2.3. Field introduced in 18.2.1.
	// Read Only: true
	SpMetadata *string `json:"sp_metadata,omitempty"`
}

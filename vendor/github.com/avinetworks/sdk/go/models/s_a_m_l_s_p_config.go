package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SAMLSPConfig s a m l s p config
// swagger:model SAMLSPConfig
type SAMLSPConfig struct {

	// HTTP cookie name for authenticated session. Field introduced in 18.2.3.
	CookieName *string `json:"cookie_name,omitempty"`

	// Cookie timeout in minutes. Allowed values are 1-1440. Field introduced in 18.2.3.
	CookieTimeout *int32 `json:"cookie_timeout,omitempty"`

	// Globally unique SAML entityID for this node. The SAML application entity ID on the IDP should match this. Field introduced in 18.2.3.
	// Required: true
	EntityID *string `json:"entity_id"`

	// Key to generate the cookie. Field introduced in 18.2.3.
	Key []*HTTPCookiePersistenceKey `json:"key,omitempty"`

	// SP will use this SSL certificate to sign assertions going to the IdP. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 18.2.3.
	SigningSslKeyAndCertificateRef *string `json:"signing_ssl_key_and_certificate_ref,omitempty"`

	// SAML Single Signon URL to be programmed on the IDP. Field introduced in 18.2.3.
	// Required: true
	SingleSignonURL *string `json:"single_signon_url"`

	// SAML SP metadata for this application. Field introduced in 18.2.3.
	// Read Only: true
	SpMetadata *string `json:"sp_metadata,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SamlServiceProviderNode saml service provider node
// swagger:model SamlServiceProviderNode
type SamlServiceProviderNode struct {

	// Globally unique entityID for this node. Entity ID on the IDP should match this. Field introduced in 17.2.3.
	EntityID *string `json:"entity_id,omitempty"`

	// Refers to the Cluster name identifier (Virtual IP or FQDN). Field introduced in 17.2.3.
	// Required: true
	Name *string `json:"name"`

	// Service Provider signing certificate for metadata. Field deprecated in 18.2.1. Field introduced in 17.2.3.
	SigningCert *string `json:"signing_cert,omitempty"`

	// Service Provider signing key for metadata. Field deprecated in 18.2.1. Field introduced in 17.2.3.
	SigningKey *string `json:"signing_key,omitempty"`

	// Service Engines will use this SSL certificate to sign assertions going to the IdP. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 18.2.1.
	SigningSslKeyAndCertificateRef *string `json:"signing_ssl_key_and_certificate_ref,omitempty"`

	// Single Signon URL to be programmed on the IDP. Field introduced in 17.2.3.
	SingleSignonURL *string `json:"single_signon_url,omitempty"`
}

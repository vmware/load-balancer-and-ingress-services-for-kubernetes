package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TCPApplicationProfile TCP application profile
// swagger:model TCPApplicationProfile
type TCPApplicationProfile struct {

	// Select the PKI profile to be associated with the Virtual Service. This profile defines the Certificate Authority and Revocation List. It is a reference to an object of type PKIProfile. Field introduced in 18.2.3. Allowed in Basic edition, Essentials edition, Enterprise edition.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// Enable/Disable the usage of proxy protocol to convey client connection information to the back-end servers.  Valid only for L4 application profiles and TCP proxy. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	ProxyProtocolEnabled *bool `json:"proxy_protocol_enabled,omitempty"`

	// Version of proxy protocol to be used to convey client connection information to the back-end servers. Enum options - PROXY_PROTOCOL_VERSION_1, PROXY_PROTOCOL_VERSION_2. Allowed in Basic(Allowed values- PROXY_PROTOCOL_VERSION_1) edition, Essentials(Allowed values- PROXY_PROTOCOL_VERSION_1) edition, Enterprise edition.
	ProxyProtocolVersion *string `json:"proxy_protocol_version,omitempty"`

	// Specifies whether the client side verification is set to none, request or require. Enum options - SSL_CLIENT_CERTIFICATE_NONE, SSL_CLIENT_CERTIFICATE_REQUEST, SSL_CLIENT_CERTIFICATE_REQUIRE. Field introduced in 18.2.3. Allowed in Basic(Allowed values- SSL_CLIENT_CERTIFICATE_NONE) edition, Essentials(Allowed values- SSL_CLIENT_CERTIFICATE_NONE) edition, Enterprise edition.
	SslClientCertificateMode *string `json:"ssl_client_certificate_mode,omitempty"`
}

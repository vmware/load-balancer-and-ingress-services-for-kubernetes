package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLProfile s s l profile
// swagger:model SSLProfile
type SSLProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Ciphers suites represented as defined by U(http //www.openssl.org/docs/apps/ciphers.html).
	AcceptedCiphers *string `json:"accepted_ciphers,omitempty"`

	// Set of versions accepted by the server.
	AcceptedVersions []*SSLVersion `json:"accepted_versions,omitempty"`

	//  Enum options - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256, TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384, TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256, TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384, TLS_RSA_WITH_AES_128_GCM_SHA256, TLS_RSA_WITH_AES_256_GCM_SHA384, TLS_RSA_WITH_AES_128_CBC_SHA256, TLS_RSA_WITH_AES_256_CBC_SHA256, TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA, TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA, TLS_RSA_WITH_AES_128_CBC_SHA, TLS_RSA_WITH_AES_256_CBC_SHA, TLS_RSA_WITH_3DES_EDE_CBC_SHA, TLS_RSA_WITH_RC4_128_SHA.
	CipherEnums []string `json:"cipher_enums,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// DH Parameters used in SSL. At this time, it is not configurable and is set to 2048 bits.
	Dhparam *string `json:"dhparam,omitempty"`

	// Enable SSL session re-use.
	EnableSslSessionReuse *bool `json:"enable_ssl_session_reuse,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Prefer the SSL cipher ordering presented by the client during the SSL handshake over the one specified in the SSL Profile.
	PreferClientCipherOrdering *bool `json:"prefer_client_cipher_ordering,omitempty"`

	// Send 'close notify' alert message for a clean shutdown of the SSL connection.
	SendCloseNotify *bool `json:"send_close_notify,omitempty"`

	// Placeholder for description of property ssl_rating of obj type SSLProfile field type str  type object
	SslRating *SSLRating `json:"ssl_rating,omitempty"`

	// The amount of time in seconds before an SSL session expires.
	SslSessionTimeout *int32 `json:"ssl_session_timeout,omitempty"`

	// Placeholder for description of property tags of obj type SSLProfile field type str  type object
	Tags []*Tag `json:"tags,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// SSL Profile Type. Enum options - SSL_PROFILE_TYPE_APPLICATION, SSL_PROFILE_TYPE_SYSTEM. Field introduced in 17.2.8.
	Type *string `json:"type,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}

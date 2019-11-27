package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorSSlattributes health monitor s slattributes
// swagger:model HealthMonitorSSLAttributes
type HealthMonitorSSlattributes struct {

	// PKI profile used to validate the SSL certificate presented by a server. It is a reference to an object of type PKIProfile. Field introduced in 17.1.1.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// Service engines will present this SSL certificate to the server. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 17.1.1.
	SslKeyAndCertificateRef *string `json:"ssl_key_and_certificate_ref,omitempty"`

	// SSL profile defines ciphers and SSL versions to be used for healthmonitor traffic to the back-end servers. It is a reference to an object of type SSLProfile. Field introduced in 17.1.1.
	// Required: true
	SslProfileRef *string `json:"ssl_profile_ref"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecureChannelConfiguration secure channel configuration
// swagger:model SecureChannelConfiguration
type SecureChannelConfiguration struct {

	// Boolean which allowed force update of secure channel certificate. Forced updating has been disallowed. Field deprecated in 18.2.8. Field introduced in 18.2.5.
	BypassSecureChannelMustChecks *bool `json:"bypass_secure_channel_must_checks,omitempty"`

	// Certificate for secure channel. Leave list empty to use system default certs. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 18.1.4, 18.2.1. Maximum of 1 items allowed.
	SslkeyandcertificateRefs []string `json:"sslkeyandcertificate_refs,omitempty"`
}

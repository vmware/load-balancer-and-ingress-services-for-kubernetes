package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLVersion s s l version
// swagger:model SSLVersion
type SSLVersion struct {

	//  Enum options - SSL_VERSION_SSLV3, SSL_VERSION_TLS1, SSL_VERSION_TLS1_1, SSL_VERSION_TLS1_2.
	// Required: true
	Type *string `json:"type"`
}

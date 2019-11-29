package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLKeyRSAParams s s l key r s a params
// swagger:model SSLKeyRSAParams
type SSLKeyRSAParams struct {

	// Number of exponent.
	Exponent *int32 `json:"exponent,omitempty"`

	//  Enum options - SSL_KEY_1024_BITS, SSL_KEY_2048_BITS, SSL_KEY_3072_BITS, SSL_KEY_4096_BITS.
	KeySize *string `json:"key_size,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPCookiePersistenceKey Http cookie persistence key
// swagger:model HttpCookiePersistenceKey
type HTTPCookiePersistenceKey struct {

	// aes_key of HttpCookiePersistenceKey.
	AesKey *string `json:"aes_key,omitempty"`

	// hmac_key of HttpCookiePersistenceKey.
	HmacKey *string `json:"hmac_key,omitempty"`

	// name to use for cookie encryption.
	Name *string `json:"name,omitempty"`
}

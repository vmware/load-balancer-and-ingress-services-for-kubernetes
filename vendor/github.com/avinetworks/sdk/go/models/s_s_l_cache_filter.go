package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLCacheFilter s s l cache filter
// swagger:model SSLCacheFilter
type SSLCacheFilter struct {

	// Hexadecimal representation of the SSL session ID. Field introduced in 20.1.1.
	SslSessionID *string `json:"ssl_session_id,omitempty"`
}

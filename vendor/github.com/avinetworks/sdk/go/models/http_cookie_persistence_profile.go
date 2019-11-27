package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPCookiePersistenceProfile Http cookie persistence profile
// swagger:model HttpCookiePersistenceProfile
type HTTPCookiePersistenceProfile struct {

	// If no persistence cookie was received from the client, always send it.
	AlwaysSendCookie *bool `json:"always_send_cookie,omitempty"`

	// HTTP cookie name for cookie persistence.
	CookieName *string `json:"cookie_name,omitempty"`

	// Key name to use for cookie encryption.
	EncryptionKey *string `json:"encryption_key,omitempty"`

	// Placeholder for description of property key of obj type HttpCookiePersistenceProfile field type str  type object
	Key []*HTTPCookiePersistenceKey `json:"key,omitempty"`

	// The length of time after a client's connections have closed before expiring the client's persistence to a server. No value or 'zero' indicates no timeout. Allowed values are 1-14400. Special values are 0- 'No Timeout'.
	Timeout *int32 `json:"timeout,omitempty"`
}

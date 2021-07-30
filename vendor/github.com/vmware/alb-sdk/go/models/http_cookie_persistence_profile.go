// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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

	// The maximum lifetime of any session cookie. No value or 'zero' indicates no timeout. Allowed values are 1-14400. Special values are 0- 'No Timeout'. Unit is MIN.
	Timeout *int32 `json:"timeout,omitempty"`
}

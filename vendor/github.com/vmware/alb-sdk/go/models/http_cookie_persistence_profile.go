// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPCookiePersistenceProfile Http cookie persistence profile
// swagger:model HttpCookiePersistenceProfile
type HTTPCookiePersistenceProfile struct {

	// If no persistence cookie was received from the client, always send it. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AlwaysSendCookie *bool `json:"always_send_cookie,omitempty"`

	// HTTP cookie name for cookie persistence. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CookieName *string `json:"cookie_name,omitempty"`

	// Key name to use for cookie encryption. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EncryptionKey *string `json:"encryption_key,omitempty"`

	// Sets the HttpOnly attribute in the cookie. Setting this helps to prevent the client side scripts from accessing this cookie, if supported by browser. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPOnly *bool `json:"http_only,omitempty"`

	// When True, the cookie used is a persistent cookie, i.e. the cookie shouldn't be used at the end of the timeout. By default, it is set to false, making the cookie a session cookie, which allows clients to use it even after the timeout, if the session is still open. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IsPersistentCookie *bool `json:"is_persistent_cookie,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Key []*HTTPCookiePersistenceKey `json:"key,omitempty"`

	// The maximum lifetime of any session cookie. No value or 'zero' indicates no timeout. Allowed values are 1-14400. Special values are 0- No Timeout. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Timeout *int32 `json:"timeout,omitempty"`
}

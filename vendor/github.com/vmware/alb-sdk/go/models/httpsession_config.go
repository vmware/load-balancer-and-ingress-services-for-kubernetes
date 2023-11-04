// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpsessionConfig httpsession config
// swagger:model HTTPSessionConfig
type HttpsessionConfig struct {

	// If set, HTTP session cookie will use 'HttpOnly' attribute. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionCookieHttponly *bool `json:"session_cookie_httponly,omitempty"`

	// HTTP session cookie name to use. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionCookieName *string `json:"session_cookie_name,omitempty"`

	// HTTP session cookie SameSite attribute. Enum options - SAMESITE_NONE, SAMESITE_LAX, SAMESITE_STRICT. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionCookieSamesite *string `json:"session_cookie_samesite,omitempty"`

	// If set, HTTP session cookie will use 'Secure' attribute. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionCookieSecure *bool `json:"session_cookie_secure,omitempty"`
}

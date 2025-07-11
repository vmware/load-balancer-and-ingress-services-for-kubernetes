// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

	// Maximum allowed time between creating a session and the client coming back. Value in seconds. Allowed values are 120-3600. Field introduced in 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionEstablishmentTimeout *uint32 `json:"session_establishment_timeout,omitempty"`

	// Maximum allowed time to expire the session after establishment on client inactivity. Value in seconds. Allowed values are 120-604800. Field introduced in 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionIDLETimeout *uint32 `json:"session_idle_timeout,omitempty"`

	// Maximum allowed time to expire the session, even if it is still active. Value in seconds. Allowed values are 120-604800. Field introduced in 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionMaximumTimeout *uint32 `json:"session_maximum_timeout,omitempty"`
}

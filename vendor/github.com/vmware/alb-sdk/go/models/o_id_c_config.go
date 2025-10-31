// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OIDCConfig o ID c config
// swagger:model OIDCConfig
type OIDCConfig struct {

	// Adds openid as one of the scopes enabling OpenID Connect flow. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OidcEnable *bool `json:"oidc_enable,omitempty"`

	// Fetch profile information by enabling profile scope. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Profile *bool `json:"profile,omitempty"`

	// Fetch profile information from Userinfo Endpoint. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Userinfo *bool `json:"userinfo,omitempty"`
}

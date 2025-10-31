// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OAuthAppSettings o auth app settings
// swagger:model OAuthAppSettings
type OAuthAppSettings struct {

	// Application specific identifier. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ClientID *string `json:"client_id"`

	// Application specific identifier secret. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientSecret *string `json:"client_secret,omitempty"`

	// OpenID Connect specific configuration. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OidcConfig *OIDCConfig `json:"oidc_config,omitempty"`

	// Scope specified to give limited access to the app. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Scopes []string `json:"scopes,omitempty"`
}

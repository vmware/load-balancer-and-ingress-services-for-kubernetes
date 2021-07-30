// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlSettings saml settings
// swagger:model SamlSettings
type SamlSettings struct {

	// Configure remote Identity provider settings. Field introduced in 17.2.3.
	Idp *SamlIdentityProviderSettings `json:"idp,omitempty"`

	// Configure service provider settings for the Controller. Field introduced in 17.2.3.
	// Required: true
	Sp *SamlServiceProviderSettings `json:"sp"`
}

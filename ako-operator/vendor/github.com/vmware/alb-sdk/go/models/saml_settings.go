// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlSettings saml settings
// swagger:model SamlSettings
type SamlSettings struct {

	// Configure remote Identity provider settings. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Idp *SamlIdentityProviderSettings `json:"idp"`

	// Configure service provider settings for the Controller. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Sp *SamlServiceProviderSettings `json:"sp"`
}

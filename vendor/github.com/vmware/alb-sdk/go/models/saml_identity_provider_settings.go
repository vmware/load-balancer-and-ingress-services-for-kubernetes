// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlIdentityProviderSettings saml identity provider settings
// swagger:model SamlIdentityProviderSettings
type SamlIdentityProviderSettings struct {

	// SAML IDP metadata. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Metadata *string `json:"metadata,omitempty"`

	// SAML IDP Federation Metadata Url. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MetadataURL *string `json:"metadata_url,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlIdentityProviderSettings saml identity provider settings
// swagger:model SamlIdentityProviderSettings
type SamlIdentityProviderSettings struct {

	// The interval to query and download SAML IDP metadata using the metadata URL. Allowed values are 1-10080. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MetaDataDownloadInterval *int32 `json:"meta_data_download_interval,omitempty"`

	// SAML IDP metadata. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Metadata *string `json:"metadata,omitempty"`

	// SAML IDP Federation Metadata Url. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MetadataURL *string `json:"metadata_url,omitempty"`

	// Enable Periodic Metadata Download. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PeriodicDownload *bool `json:"periodic_download,omitempty"`
}

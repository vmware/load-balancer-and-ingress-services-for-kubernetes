// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecureChannelToken secure channel token
// swagger:model SecureChannelToken
type SecureChannelToken struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Expiry time for auth_token. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExpiryTime *float64 `json:"expiry_time,omitempty"`

	// Whether this auth_token is used by some node(SE/controller). Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	InUse *bool `json:"in_use,omitempty"`

	// Metadata associated with auth_token. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Metadata []*SecureChannelMetadata `json:"metadata,omitempty"`

	// Auth_token used for SE/controller authorization. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Auth_token used for SE/controller authorization. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

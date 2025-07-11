// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AppCookiePersistenceProfile app cookie persistence profile
// swagger:model AppCookiePersistenceProfile
type AppCookiePersistenceProfile struct {

	// Key to use for cookie encryption. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EncryptionKey *string `json:"encryption_key,omitempty"`

	// Header or cookie name for application cookie persistence. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	PrstHdrName *string `json:"prst_hdr_name"`

	// The length of time after a client's connections have closed before expiring the client's persistence to a server. Allowed values are 1-720. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Timeout *int32 `json:"timeout,omitempty"`
}

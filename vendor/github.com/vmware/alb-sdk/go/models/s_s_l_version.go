// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLVersion s s l version
// swagger:model SSLVersion
type SSLVersion struct {

	//  Enum options - SSL_VERSION_SSLV3, SSL_VERSION_TLS1, SSL_VERSION_TLS1_1, SSL_VERSION_TLS1_2, SSL_VERSION_TLS1_3. Allowed in Basic(Allowed values- SSL_VERSION_SSLV3,SSL_VERSION_TLS1,SSL_VERSION_TLS1_1,SSL_VERSION_TLS1_2) edition, Essentials(Allowed values- SSL_VERSION_SSLV3,SSL_VERSION_TLS1,SSL_VERSION_TLS1_1,SSL_VERSION_TLS1_2) edition, Enterprise edition.
	// Required: true
	Type *string `json:"type"`
}

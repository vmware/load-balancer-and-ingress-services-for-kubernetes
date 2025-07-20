// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecureChannelConfiguration secure channel configuration
// swagger:model SecureChannelConfiguration
type SecureChannelConfiguration struct {

	// Certificate for secure channel. Leave list empty to use system default certs. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 18.1.4, 18.2.1. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslkeyandcertificateRefs []string `json:"sslkeyandcertificate_refs,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OCICredentials o c i credentials
// swagger:model OCICredentials
type OCICredentials struct {

	// API key with respect to the Public Key. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fingerprint *string `json:"fingerprint,omitempty"`

	// Private Key file (pem file) content. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeyContent *string `json:"key_content,omitempty"`

	// Pass phrase for the key. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PassPhrase *string `json:"pass_phrase,omitempty"`

	// Oracle Cloud Id for the User. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	User *string `json:"user,omitempty"`
}

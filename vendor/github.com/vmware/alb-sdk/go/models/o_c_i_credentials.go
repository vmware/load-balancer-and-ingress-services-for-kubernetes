// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OCICredentials o c i credentials
// swagger:model OCICredentials
type OCICredentials struct {

	// API key with respect to the Public Key. Field introduced in 18.2.1,18.1.3.
	Fingerprint *string `json:"fingerprint,omitempty"`

	// Private Key file (pem file) content. Field introduced in 18.2.1,18.1.3.
	KeyContent *string `json:"key_content,omitempty"`

	// Pass phrase for the key. Field introduced in 18.2.1,18.1.3.
	PassPhrase *string `json:"pass_phrase,omitempty"`

	// Oracle Cloud Id for the User. Field introduced in 18.2.1,18.1.3.
	User *string `json:"user,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JWSKey j w s key
// swagger:model JWSKey
type JWSKey struct {

	// Algorithm that need to be used while signing/validation, allowed values  HS256, HS384, HS512. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Alg *string `json:"alg,omitempty"`

	// Secret JWK for signing/validation, length of the key varies depending upon the type of algorithm used for key generation {HS256  32 bytes, HS384  48bytes, HS512  64 bytes}. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Key *string `json:"key"`

	// Unique key id across all keys. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Kid *string `json:"kid"`

	// Secret key type/format, allowed value  octet(oct). Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Kty *string `json:"kty,omitempty"`
}

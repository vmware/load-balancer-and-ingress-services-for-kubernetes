// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPCookiePersistenceKey Http cookie persistence key
// swagger:model HttpCookiePersistenceKey
type HTTPCookiePersistenceKey struct {

	// aes_key of HttpCookiePersistenceKey.
	AesKey *string `json:"aes_key,omitempty"`

	// hmac_key of HttpCookiePersistenceKey.
	HmacKey *string `json:"hmac_key,omitempty"`

	// name to use for cookie encryption.
	Name *string `json:"name,omitempty"`
}

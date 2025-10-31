// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLKeyECParams s s l key e c params
// swagger:model SSLKeyECParams
type SSLKeyECParams struct {

	//  Enum options - SSL_KEY_EC_CURVE_SECP256R1, SSL_KEY_EC_CURVE_SECP384R1, SSL_KEY_EC_CURVE_SECP521R1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Curve *string `json:"curve,omitempty"`
}

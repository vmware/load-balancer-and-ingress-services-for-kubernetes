// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLKeyParams s s l key params
// swagger:model SSLKeyParams
type SSLKeyParams struct {

	//  Enum options - SSL_KEY_ALGORITHM_RSA, SSL_KEY_ALGORITHM_EC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Algorithm *string `json:"algorithm"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EcParams *SSLKeyECParams `json:"ec_params,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RsaParams *SSLKeyRSAParams `json:"rsa_params,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TLSTicket TLS ticket
// swagger:model TLSTicket
type TLSTicket struct {

	// aes_key of TLSTicket.
	// Required: true
	AesKey *string `json:"aes_key"`

	// hmac_key of TLSTicket.
	// Required: true
	HmacKey *string `json:"hmac_key"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`
}

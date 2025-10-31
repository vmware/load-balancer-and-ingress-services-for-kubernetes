// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NTPAuthenticationKey n t p authentication key
// swagger:model NTPAuthenticationKey
type NTPAuthenticationKey struct {

	// Message Digest Algorithm used for NTP authentication. Default is NTP_AUTH_ALGORITHM_MD5. Enum options - NTP_AUTH_ALGORITHM_MD5, NTP_AUTH_ALGORITHM_SHA1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Algorithm *string `json:"algorithm,omitempty"`

	// NTP Authentication key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Key *string `json:"key"`

	// Key number to be assigned to the authentication-key. Allowed values are 1-65534. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	KeyNumber *uint32 `json:"key_number"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLCacheFilter s s l cache filter
// swagger:model SSLCacheFilter
type SSLCacheFilter struct {

	// Hexadecimal representation of the SSL session ID. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslSessionID *string `json:"ssl_session_id,omitempty"`
}

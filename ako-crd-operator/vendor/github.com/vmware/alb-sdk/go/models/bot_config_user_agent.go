// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotConfigUserAgent bot config user agent
// swagger:model BotConfigUserAgent
type BotConfigUserAgent struct {

	// Whether User Agent-based Bot detection is enabled. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Whether to match the TLS fingerprint observed on the request against TLS fingerprints expected for the user agent. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UseTLSFingerprint *bool `json:"use_tls_fingerprint,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClientFingerprints client fingerprints
// swagger:model ClientFingerprints
type ClientFingerprints struct {

	// Message Digest (md5) of filtered JA3 from ClientHello. This can deviate from 'tls_fingerprint' because not all extensions are considered. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FilteredTLSFingerprint *string `json:"filtered_tls_fingerprint,omitempty"`

	// Message Digest (md5) of JA3 from ClientHello. Only present if the full TLS fingerprint is different from the filtered fingerprint. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FullTLSFingerprint *string `json:"full_tls_fingerprint,omitempty"`

	// Message Digest (md5) of normalized JA3 from ClientHello. This can deviate from 'full_tls_fingerprint' because extensions 21 and 35 are removed and the remaining values are sorted numerically before the MD5 is calculated. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NormalizedTLSFingerprint *string `json:"normalized_tls_fingerprint,omitempty"`

	// Values of selected fields from the ClientHello. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TLSClientInfo *TLSClientInfo `json:"tls_client_info,omitempty"`
}

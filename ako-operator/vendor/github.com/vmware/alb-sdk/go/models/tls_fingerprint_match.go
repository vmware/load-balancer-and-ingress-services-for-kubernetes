// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TLSFingerprintMatch Tls fingerprint match
// swagger:model TlsFingerprintMatch
type TLSFingerprintMatch struct {

	// The list of fingerprints. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Fingerprints []string `json:"fingerprints,omitempty"`

	// Match criteria. Enum options - IS_IN, IS_NOT_IN. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MatchOperation *string `json:"match_operation"`

	// UUIDs of the *string groups. It is a reference to an object of type StringGroup. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}

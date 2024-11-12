// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CRL c r l
// swagger:model CRL
type CRL struct {

	// Common name of the issuer in the Certificate Revocation list. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CommonName *string `json:"common_name,omitempty"`

	// Distinguished name of the issuer in the Certificate Revocation list. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DistinguishedName *string `json:"distinguished_name,omitempty"`

	// Cached etag to optimize the download of the CRL. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Etag *string `json:"etag,omitempty"`

	// Fingerprint of the CRL. Used to avoid configuring duplicates. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fingerprint *string `json:"fingerprint,omitempty"`

	// Last time CRL was refreshed by the system. This is an internal field used by the system. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastRefreshed *string `json:"last_refreshed,omitempty"`

	// The date when this CRL was last issued. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastUpdate *string `json:"last_update,omitempty"`

	// The date when a newer CRL will be available. Also conveys the date after which the CRL should be considered obsolete. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NextUpdate *string `json:"next_update,omitempty"`

	// URL of a server that issues the Certificate Revocation list. If this is configured, CRL will be periodically downloaded either based on the configured update interval or the next update interval in the CRL. CRL itself is stored in the body. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerURL *string `json:"server_url,omitempty"`

	// Certificate Revocation list in plain text for readability. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Text *string `json:"text,omitempty"`

	// Interval in minutes to check for CRL update. If not specified, interval will be 1 day. Allowed values are 30-525600. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpdateInterval *int32 `json:"update_interval,omitempty"`
}

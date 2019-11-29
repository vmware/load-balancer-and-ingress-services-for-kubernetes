package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CRL c r l
// swagger:model CRL
type CRL struct {

	// Certificate Revocation list from a given issuer in PEM format. This can either be configured directly or via the server_url. .
	Body *string `json:"body,omitempty"`

	// Common name of the issuer in the Certificate Revocation list.
	CommonName *string `json:"common_name,omitempty"`

	// Distinguished name of the issuer in the Certificate Revocation list.
	DistinguishedName *string `json:"distinguished_name,omitempty"`

	// Cached etag to optimize the download of the CRL.
	Etag *string `json:"etag,omitempty"`

	// Fingerprint of the CRL. Used to avoid configuring duplicates.
	Fingerprint *string `json:"fingerprint,omitempty"`

	// Last time CRL was refreshed by the system. This is an internal field used by the system.
	LastRefreshed *string `json:"last_refreshed,omitempty"`

	// The date when this CRL was last issued.
	LastUpdate *string `json:"last_update,omitempty"`

	// The date when a newer CRL will be available. Also conveys the date after which the CRL should be considered obsolete.
	NextUpdate *string `json:"next_update,omitempty"`

	// URL of a server that issues the Certificate Revocation list. If this is configured, CRL will be periodically downloaded either based on the configured update interval or the next update interval in the CRL. CRL itself is stored in the body.
	ServerURL *string `json:"server_url,omitempty"`

	// Certificate Revocation list in plain text for readability.
	Text *string `json:"text,omitempty"`

	// Interval in minutes to check for CRL update. If not specified, interval will be 1 day. Allowed values are 30-MAX.
	UpdateInterval *int32 `json:"update_interval,omitempty"`
}

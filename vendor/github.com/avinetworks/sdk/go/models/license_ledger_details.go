package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LicenseLedgerDetails license ledger details
// swagger:model LicenseLedgerDetails
type LicenseLedgerDetails struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Maintain information about reservation against cookie. Field introduced in 20.1.1.
	EscrowInfos []*LicenseInfo `json:"escrow_infos,omitempty"`

	// Maintain information about consumed licenses against se_uuid. Field introduced in 20.1.1.
	SeInfos []*LicenseInfo `json:"se_infos,omitempty"`

	// License usage per tier. Field introduced in 20.1.1.
	TierUsages []*LicenseTierUsage `json:"tier_usages,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Uuid for reference. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LicenseTierSwitchDetiails license tier switch detiails
// swagger:model LicenseTierSwitchDetiails
type LicenseTierSwitchDetiails struct {

	// destination_tier of LicenseTierSwitchDetiails.
	DestinationTier *string `json:"destination_tier,omitempty"`

	// reason of LicenseTierSwitchDetiails.
	Reason *string `json:"reason,omitempty"`

	// source_tier of LicenseTierSwitchDetiails.
	SourceTier *string `json:"source_tier,omitempty"`

	// status of LicenseTierSwitchDetiails.
	Status *string `json:"status,omitempty"`
}

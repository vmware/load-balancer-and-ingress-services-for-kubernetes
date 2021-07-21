// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

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

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeUpgradeVsDisruptedEventDetails se upgrade vs disrupted event details
// swagger:model SeUpgradeVsDisruptedEventDetails
type SeUpgradeVsDisruptedEventDetails struct {

	// ip of SeUpgradeVsDisruptedEventDetails.
	IP *string `json:"ip,omitempty"`

	// notes of SeUpgradeVsDisruptedEventDetails.
	Notes []string `json:"notes,omitempty"`

	// vip_id of SeUpgradeVsDisruptedEventDetails.
	VipID *string `json:"vip_id,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}

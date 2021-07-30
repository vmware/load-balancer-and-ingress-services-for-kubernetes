// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SwitchoverFailEventDetails switchover fail event details
// swagger:model SwitchoverFailEventDetails
type SwitchoverFailEventDetails struct {

	// from_se_name of SwitchoverFailEventDetails.
	FromSeName *string `json:"from_se_name,omitempty"`

	// ip of SwitchoverFailEventDetails.
	IP *string `json:"ip,omitempty"`

	// ip6 of SwitchoverFailEventDetails.
	Ip6 *string `json:"ip6,omitempty"`

	// reason of SwitchoverFailEventDetails.
	Reason *string `json:"reason,omitempty"`

	// vs_name of SwitchoverFailEventDetails.
	VsName *string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`
}

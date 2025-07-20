// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHmEventPoolDetails se hm event pool details
// swagger:model SeHmEventPoolDetails
type SeHmEventPoolDetails struct {

	// HA Compromised reason. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HaReason *string `json:"ha_reason,omitempty"`

	// Percentage of servers up. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PercentServersUp *string `json:"percent_servers_up,omitempty"`

	// Pool name. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Pool *string `json:"pool,omitempty"`

	// Service Engine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// Server details. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Server *SeHmEventServerDetails `json:"server,omitempty"`

	// UUID of the event generator. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SrcUUID *string `json:"src_uuid,omitempty"`

	// Virtual service name. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualService *string `json:"virtual_service,omitempty"`
}

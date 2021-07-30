// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHmEventPoolDetails se hm event pool details
// swagger:model SeHmEventPoolDetails
type SeHmEventPoolDetails struct {

	// HA Compromised reason.
	HaReason *string `json:"ha_reason,omitempty"`

	// Percentage of servers up.
	PercentServersUp *string `json:"percent_servers_up,omitempty"`

	// Pool name. It is a reference to an object of type Pool.
	Pool *string `json:"pool,omitempty"`

	// Service Engine.
	SeName *string `json:"se_name,omitempty"`

	// Server details.
	Server *SeHmEventServerDetails `json:"server,omitempty"`

	// UUID of the event generator.
	SrcUUID *string `json:"src_uuid,omitempty"`

	// Virtual service name. It is a reference to an object of type VirtualService.
	VirtualService *string `json:"virtual_service,omitempty"`
}

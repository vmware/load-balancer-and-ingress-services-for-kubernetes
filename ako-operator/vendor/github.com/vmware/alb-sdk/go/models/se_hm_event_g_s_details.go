// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHmEventGSDetails se hm event g s details
// swagger:model SeHmEventGSDetails
type SeHmEventGSDetails struct {

	// GslbService name. It is a reference to an object of type GslbService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbService *string `json:"gslb_service,omitempty"`

	// HA Compromised reason. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HaReason *string `json:"ha_reason,omitempty"`

	// Reason Gslb Service is down. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Service Engine name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the event generator. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SrcUUID *string `json:"src_uuid,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHmEventVsDetails se hm event vs details
// swagger:model SeHmEventVsDetails
type SeHmEventVsDetails struct {

	// HA Compromised reason. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HaReason *string `json:"ha_reason,omitempty"`

	// Reason for Virtual Service Down. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Service Engine name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the event generator. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SrcUUID *string `json:"src_uuid,omitempty"`

	// VIP address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vip6Address *IPAddr `json:"vip6_address,omitempty"`

	// VIP address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipAddress *IPAddr `json:"vip_address,omitempty"`

	// VIP id. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipID *string `json:"vip_id,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualService *string `json:"virtual_service,omitempty"`
}

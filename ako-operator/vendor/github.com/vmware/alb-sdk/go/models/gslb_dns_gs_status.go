// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbDNSGsStatus gslb Dns gs status
// swagger:model GslbDnsGsStatus
type GslbDNSGsStatus struct {

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Counter to track the number of partial updates sent.  Once it reaches the partial updates threshold, a full update is sent. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumPartialUpdates *uint32 `json:"num_partial_updates,omitempty"`

	// Threshold after which a full GS Status is sent. . Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PartialUpdateThreshold *uint32 `json:"partial_update_threshold,omitempty"`

	// State variable to trigger full or partial update. Enum options - GSLB_FULL_UPDATE_PENDING, GSLB_PARTIAL_UPDATE_PENDING. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`

	// Describes the type (partial/full) of the last GS status sent to Dns-VS(es). Enum options - GSLB_NO_UPDATE, GSLB_FULL_UPDATE, GSLB_PARTIAL_UPDATE. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`
}

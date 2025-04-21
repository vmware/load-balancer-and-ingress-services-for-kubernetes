// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeLicensedBandwdithExceededEventDetails se licensed bandwdith exceeded event details
// swagger:model SeLicensedBandwdithExceededEventDetails
type SeLicensedBandwdithExceededEventDetails struct {

	// Number of packets dropped since the last event. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumPktsDropped uint32 `json:"num_pkts_dropped,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`
}

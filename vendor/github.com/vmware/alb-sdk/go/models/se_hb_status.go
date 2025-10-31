// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHbStatus se hb status
// swagger:model SeHbStatus
type SeHbStatus struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	LastHbReqSent *string `json:"last_hb_req_sent"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	LastHbRespRecv *string `json:"last_hb_resp_recv"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NumHbMisses *int32 `json:"num_hb_misses"`

	//  Field introduced in 17.2.10,18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOutstandingHb *int32 `json:"num_outstanding_hb,omitempty"`
}

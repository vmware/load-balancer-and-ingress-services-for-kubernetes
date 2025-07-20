// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SePoolLbEventDetails se pool lb event details
// swagger:model SePoolLbEventDetails
type SePoolLbEventDetails struct {

	// Reason code for load balancing failure. Enum options - PERSISTENT_SERVER_INVALID, PERSISTENT_SERVER_DOWN, SRVR_DOWN, ADD_PENDING, SLOW_START_MAX_CONN, MAX_CONN, NO_LPORT, SUSPECT_STATE, MAX_CONN_RATE, CAPEST_RAND_MAX_CONN, GET_NEXT, PERSISTENT_SERVER_DISABLED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FailureCode *string `json:"failure_code,omitempty"`

	// Pool name. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Pool *string `json:"pool,omitempty"`

	// Reason for Load balancing failure. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// UUID of event generator. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SrcUUID *string `json:"src_uuid,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualService *string `json:"virtual_service,omitempty"`
}

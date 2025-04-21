// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorSctp health monitor sctp
// swagger:model HealthMonitorSctp
type HealthMonitorSctp struct {

	// Request data to send after completing the SCTP handshake. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SctpRequest *string `json:"sctp_request,omitempty"`

	// Match for the desired keyword in the first 2Kb of the server's SCTP response. If this field is left blank, no server response is required. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SctpResponse *string `json:"sctp_response,omitempty"`
}

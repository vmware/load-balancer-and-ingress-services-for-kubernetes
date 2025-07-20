// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorTCP health monitor Tcp
// swagger:model HealthMonitorTcp
type HealthMonitorTCP struct {

	// Match or look for this keyword in the first 2KB of server's response indicating server maintenance.  A successful match results in the server being marked down. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaintenanceResponse *string `json:"maintenance_response,omitempty"`

	// Configure TCP health monitor to use half-open TCP connections to monitor the health of backend servers thereby avoiding consumption of a full fledged server side connection and the overhead and logs associated with it.  This method is light-weight as it makes use of listener in server's kernel layer to measure the health and a child socket or user thread is not created on the server side. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	TCPHalfOpen *bool `json:"tcp_half_open,omitempty"`

	// Request data to send after completing the TCP handshake. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPRequest *string `json:"tcp_request,omitempty"`

	// Match for the desired keyword in the first 2Kb of the server's TCP response. If this field is left blank, no server response is required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPResponse *string `json:"tcp_response,omitempty"`
}

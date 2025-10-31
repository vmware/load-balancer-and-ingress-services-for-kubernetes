// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorUDP health monitor Udp
// swagger:model HealthMonitorUdp
type HealthMonitorUDP struct {

	// Match or look for this keyword in the first 2KB of server's response indicating server maintenance.  A successful match results in the server being marked down. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaintenanceResponse *string `json:"maintenance_response,omitempty"`

	// Send UDP request. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UDPRequest *string `json:"udp_request,omitempty"`

	// Match for keyword in the UDP response. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UDPResponse *string `json:"udp_response,omitempty"`
}

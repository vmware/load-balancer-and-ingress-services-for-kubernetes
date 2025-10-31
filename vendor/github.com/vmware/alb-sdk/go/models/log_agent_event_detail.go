// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LogAgentEventDetail log agent event detail
// swagger:model LogAgentEventDetail
type LogAgentEventDetail struct {

	// Protocol used for communication to the external entity. Enum options - TCP_CONN. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Protocol *string `json:"protocol"`

	// Event for TCP connection restablishment rate exceeds configured threshold. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RapidConnection *LogAgentTCPConnEstRateExcdEvent `json:"rapid_connection,omitempty"`

	// Event details for TCP connection event. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TCPDetail *LogAgentTCPClientEventDetail `json:"tcp_detail,omitempty"`

	// Type of log agent event. Enum options - LOG_AGENT_CONNECTION_ERROR. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}

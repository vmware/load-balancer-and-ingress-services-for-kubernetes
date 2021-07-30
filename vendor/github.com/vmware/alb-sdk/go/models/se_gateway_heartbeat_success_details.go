// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeGatewayHeartbeatSuccessDetails se gateway heartbeat success details
// swagger:model SeGatewayHeartbeatSuccessDetails
type SeGatewayHeartbeatSuccessDetails struct {

	// IP address of gateway monitored.
	// Required: true
	GatewayIP *string `json:"gateway_ip"`

	// Name of Virtual Routing Context in which this gateway is present.
	VrfName *string `json:"vrf_name,omitempty"`

	// UUID of the Virtual Routing Context in which this gateway is present.
	VrfUUID *string `json:"vrf_uuid,omitempty"`
}

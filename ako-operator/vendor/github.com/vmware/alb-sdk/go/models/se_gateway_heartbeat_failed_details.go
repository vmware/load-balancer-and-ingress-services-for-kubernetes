// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeGatewayHeartbeatFailedDetails se gateway heartbeat failed details
// swagger:model SeGatewayHeartbeatFailedDetails
type SeGatewayHeartbeatFailedDetails struct {

	// IP address of gateway monitored. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	GatewayIP *string `json:"gateway_ip"`

	// Name of Virtual Routing Context in which this gateway is present. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfName *string `json:"vrf_name,omitempty"`

	// UUID of the Virtual Routing Context in which this gateway is present. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfUUID *string `json:"vrf_uuid,omitempty"`
}

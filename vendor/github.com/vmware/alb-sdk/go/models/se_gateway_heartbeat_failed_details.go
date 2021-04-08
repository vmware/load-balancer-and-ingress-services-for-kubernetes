package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeGatewayHeartbeatFailedDetails se gateway heartbeat failed details
// swagger:model SeGatewayHeartbeatFailedDetails
type SeGatewayHeartbeatFailedDetails struct {

	// IP address of gateway monitored.
	// Required: true
	GatewayIP *string `json:"gateway_ip"`

	// Name of Virtual Routing Context in which this gateway is present.
	VrfName *string `json:"vrf_name,omitempty"`

	// UUID of the Virtual Routing Context in which this gateway is present.
	VrfUUID *string `json:"vrf_uuid,omitempty"`
}

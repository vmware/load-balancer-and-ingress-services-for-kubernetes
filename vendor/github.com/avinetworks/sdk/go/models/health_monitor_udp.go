package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorUDP health monitor Udp
// swagger:model HealthMonitorUdp
type HealthMonitorUDP struct {

	// Match or look for this keyword in the first 2KB of server's response indicating server maintenance.  A successful match results in the server being marked down.
	MaintenanceResponse *string `json:"maintenance_response,omitempty"`

	// Send UDP request.
	UDPRequest *string `json:"udp_request,omitempty"`

	// Match for keyword in the UDP response.
	UDPResponse *string `json:"udp_response,omitempty"`
}

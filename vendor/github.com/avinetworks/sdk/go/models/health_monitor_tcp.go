package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorTCP health monitor Tcp
// swagger:model HealthMonitorTcp
type HealthMonitorTCP struct {

	// Match or look for this keyword in the first 2KB of server's response indicating server maintenance.  A successful match results in the server being marked down.
	MaintenanceResponse *string `json:"maintenance_response,omitempty"`

	// Configure TCP health monitor to use half-open TCP connections to monitor the health of backend servers thereby avoiding consumption of a full fledged server side connection and the overhead and logs associated with it.  This method is light-weight as it makes use of listener in server's kernel layer to measure the health and a child socket or user thread is not created on the server side.
	TCPHalfOpen *bool `json:"tcp_half_open,omitempty"`

	// Request data to send after completing the TCP handshake.
	TCPRequest *string `json:"tcp_request,omitempty"`

	// Match for the desired keyword in the first 2Kb of the server's TCP response. If this field is left blank, no server response is required.
	TCPResponse *string `json:"tcp_response,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitor health monitor
// swagger:model HealthMonitor
type HealthMonitor struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Placeholder for description of property dns_monitor of obj type HealthMonitor field type str  type object
	DNSMonitor *HealthMonitorDNS `json:"dns_monitor,omitempty"`

	// Placeholder for description of property external_monitor of obj type HealthMonitor field type str  type object
	ExternalMonitor *HealthMonitorExternal `json:"external_monitor,omitempty"`

	// Number of continuous failed health checks before the server is marked down. Allowed values are 1-50.
	FailedChecks *int32 `json:"failed_checks,omitempty"`

	// Placeholder for description of property http_monitor of obj type HealthMonitor field type str  type object
	HTTPMonitor *HealthMonitorHTTP `json:"http_monitor,omitempty"`

	// Placeholder for description of property https_monitor of obj type HealthMonitor field type str  type object
	HTTPSMonitor *HealthMonitorHTTP `json:"https_monitor,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster and its associated service-engines.  If the field is set to true, then the object is replicated across the federation.  . Field introduced in 17.1.3.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Use this port instead of the port defined for the server in the Pool. If the monitor succeeds to this port, the load balanced traffic will still be sent to the port of the server defined within the Pool. Allowed values are 1-65535. Special values are 0 - 'Use server port'.
	MonitorPort *int32 `json:"monitor_port,omitempty"`

	// A user friendly name for this health monitor.
	// Required: true
	Name *string `json:"name"`

	// A valid response from the server is expected within the receive timeout window.  This timeout must be less than the send interval.  If server status is regularly flapping up and down, consider increasing this value. Allowed values are 1-2400.
	ReceiveTimeout *int32 `json:"receive_timeout,omitempty"`

	// Frequency, in seconds, that monitors are sent to a server. Allowed values are 1-3600.
	SendInterval *int32 `json:"send_interval,omitempty"`

	// Health monitor for SIP. Field introduced in 17.2.8, 18.1.3, 18.2.1.
	SipMonitor *HealthMonitorSIP `json:"sip_monitor,omitempty"`

	// Number of continuous successful health checks before server is marked up. Allowed values are 1-50.
	SuccessfulChecks *int32 `json:"successful_checks,omitempty"`

	// Placeholder for description of property tcp_monitor of obj type HealthMonitor field type str  type object
	TCPMonitor *HealthMonitorTCP `json:"tcp_monitor,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Type of the health monitor. Enum options - HEALTH_MONITOR_PING, HEALTH_MONITOR_TCP, HEALTH_MONITOR_HTTP, HEALTH_MONITOR_HTTPS, HEALTH_MONITOR_EXTERNAL, HEALTH_MONITOR_UDP, HEALTH_MONITOR_DNS, HEALTH_MONITOR_GSLB, HEALTH_MONITOR_SIP.
	// Required: true
	Type *string `json:"type"`

	// Placeholder for description of property udp_monitor of obj type HealthMonitor field type str  type object
	UDPMonitor *HealthMonitorUDP `json:"udp_monitor,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the health monitor.
	UUID *string `json:"uuid,omitempty"`
}

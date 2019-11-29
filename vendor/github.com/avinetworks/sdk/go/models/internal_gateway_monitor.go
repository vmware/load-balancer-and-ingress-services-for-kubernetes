package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// InternalGatewayMonitor internal gateway monitor
// swagger:model InternalGatewayMonitor
type InternalGatewayMonitor struct {

	// Disable the gateway monitor for default gateway. They are monitored by default. Field introduced in 17.1.1.
	DisableGatewayMonitor *bool `json:"disable_gateway_monitor,omitempty"`

	// The number of consecutive failed gateway health checks before a gateway is marked down. Allowed values are 3-50. Field introduced in 17.1.1.
	GatewayMonitorFailureThreshold *int32 `json:"gateway_monitor_failure_threshold,omitempty"`

	// The interval between two ping requests sent by the gateway monitor in milliseconds. If a value is not specified, requests are sent every second. Allowed values are 100-60000. Field introduced in 17.1.1.
	GatewayMonitorInterval *int32 `json:"gateway_monitor_interval,omitempty"`

	// The number of consecutive successful gateway health checks before a gateway that was marked down by the gateway monitor is marked up. Allowed values are 3-50. Field introduced in 17.1.1.
	GatewayMonitorSuccessThreshold *int32 `json:"gateway_monitor_success_threshold,omitempty"`
}

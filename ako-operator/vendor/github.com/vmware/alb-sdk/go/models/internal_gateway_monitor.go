// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InternalGatewayMonitor internal gateway monitor
// swagger:model InternalGatewayMonitor
type InternalGatewayMonitor struct {

	// Disable the gateway monitor for default gateway. They are monitored by default. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableGatewayMonitor *bool `json:"disable_gateway_monitor,omitempty"`

	// The number of consecutive failed gateway health checks before a gateway is marked down. Allowed values are 3-50. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GatewayMonitorFailureThreshold *uint32 `json:"gateway_monitor_failure_threshold,omitempty"`

	// The interval between two ping requests sent by the gateway monitor in milliseconds. If a value is not specified, requests are sent every second. Allowed values are 100-60000. Field introduced in 17.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GatewayMonitorInterval *uint32 `json:"gateway_monitor_interval,omitempty"`

	// The number of consecutive successful gateway health checks before a gateway that was marked down by the gateway monitor is marked up. Allowed values are 3-50. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GatewayMonitorSuccessThreshold *uint32 `json:"gateway_monitor_success_threshold,omitempty"`
}

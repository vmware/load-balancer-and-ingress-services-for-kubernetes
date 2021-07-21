// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorRadius health monitor radius
// swagger:model HealthMonitorRadius
type HealthMonitorRadius struct {

	// Radius monitor will query Radius server with this password. Field introduced in 18.2.3.
	// Required: true
	Password *string `json:"password"`

	// Radius monitor will query Radius server with this shared secret. Field introduced in 18.2.3.
	// Required: true
	SharedSecret *string `json:"shared_secret"`

	// Radius monitor will query Radius server with this username. Field introduced in 18.2.3.
	// Required: true
	Username *string `json:"username"`
}

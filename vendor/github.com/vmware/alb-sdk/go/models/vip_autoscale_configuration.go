// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipAutoscaleConfiguration vip autoscale configuration
// swagger:model VipAutoscaleConfiguration
type VipAutoscaleConfiguration struct {

	// This is the list of AZ+Subnet in which Vips will be spawned. Field introduced in 17.2.12, 18.1.2.
	Zones []*VipAutoscaleZones `json:"zones,omitempty"`
}

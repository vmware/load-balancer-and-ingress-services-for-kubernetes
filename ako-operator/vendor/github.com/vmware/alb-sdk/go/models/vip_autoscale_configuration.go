// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipAutoscaleConfiguration vip autoscale configuration
// swagger:model VipAutoscaleConfiguration
type VipAutoscaleConfiguration struct {

	// This is the list of AZ+Subnet in which Vips will be spawned. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Zones []*VipAutoscaleZones `json:"zones,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipAutoscaleZones vip autoscale zones
// swagger:model VipAutoscaleZones
type VipAutoscaleZones struct {

	// Availability zone associated with the subnet. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// Determines if the subnet is capable of hosting publicly accessible IP. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	FipCapable *bool `json:"fip_capable,omitempty"`

	// UUID of the subnet for new IP address allocation. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubnetUUID *string `json:"subnet_uuid,omitempty"`
}

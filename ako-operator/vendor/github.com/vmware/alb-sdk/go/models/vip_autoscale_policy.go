// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipAutoscalePolicy vip autoscale policy
// swagger:model VipAutoscalePolicy
type VipAutoscalePolicy struct {

	// The amount of time, in seconds, when a Vip is withdrawn before a scaling activity starts. Field introduced in 17.2.12, 18.1.2. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSCooldown *uint32 `json:"dns_cooldown,omitempty"`

	// The maximum size of the group. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxSize *uint32 `json:"max_size,omitempty"`

	// The minimum size of the group. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinSize *uint32 `json:"min_size,omitempty"`

	// When set, scaling is suspended. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Suspend *bool `json:"suspend,omitempty"`
}

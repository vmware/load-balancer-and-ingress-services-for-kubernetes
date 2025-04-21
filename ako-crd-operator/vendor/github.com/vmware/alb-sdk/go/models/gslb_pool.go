// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbPool gslb pool
// swagger:model GslbPool
type GslbPool struct {

	// The load balancing algorithm will pick a local member within the GSLB service list of available Members. Enum options - GSLB_ALGORITHM_ROUND_ROBIN, GSLB_ALGORITHM_CONSISTENT_HASH, GSLB_ALGORITHM_GEO, GSLB_ALGORITHM_TOPOLOGY, GSLB_ALGORITHM_PREFERENCE_ORDER. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Algorithm *string `json:"algorithm"`

	// Mask to be applied on client IP for consistent hash algorithm. Allowed values are 1-31. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConsistentHashMask uint32 `json:"consistent_hash_mask,omitempty"`

	// Mask to be applied on client IPV6 address for consistent hash algorithm. Allowed values are 1-127. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConsistentHashMask6 uint32 `json:"consistent_hash_mask6,omitempty"`

	// User provided information that records member details such as application owner name, contact, etc. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable or disable a GSLB service pool. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// The fallback load balancing algorithm used to pick a member when the pool algorithm fails to find a valid member. For instance when algorithm is Geo and client/server do not have valid geo location. Enum options - GSLB_ALGORITHM_ROUND_ROBIN, GSLB_ALGORITHM_CONSISTENT_HASH, GSLB_ALGORITHM_GEO, GSLB_ALGORITHM_TOPOLOGY, GSLB_ALGORITHM_PREFERENCE_ORDER. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FallbackAlgorithm *string `json:"fallback_algorithm,omitempty"`

	// Manually resume traffic to a pool member once it goes down. If enabled a pool member once goes down is kept in admin down state unless admin re enables it. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ManualResume *bool `json:"manual_resume,omitempty"`

	// Select list of VIPs belonging to this GSLB service. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Members []*GslbPoolMember `json:"members,omitempty"`

	// Minimum number of health monitors in UP state to mark the member UP. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinHealthMonitorsUp uint32 `json:"min_health_monitors_up,omitempty"`

	// Name of the GSLB service pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Priority of this pool of Members. The higher the number, the higher is the priority of the pool. The DNS Service chooses the pool with the highest priority that is operationally up. Allowed values are 0-100. Special values are 0 - Do not choose members from this pool.A priority of 0 is equivalent to disabling the pool.. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Priority *uint32 `json:"priority,omitempty"`
}

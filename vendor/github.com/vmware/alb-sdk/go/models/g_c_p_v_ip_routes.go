// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPVIPRoutes g c p v IP routes
// swagger:model GCPVIPRoutes
type GCPVIPRoutes struct {

	// Match SE group subnets for VIP placement. Default is to not match SE group subnets. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchSeGroupSubnet *bool `json:"match_se_group_subnet,omitempty"`

	// Priority of the routes created in GCP. Field introduced in 20.1.7, 21.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RoutePriority *uint32 `json:"route_priority,omitempty"`
}

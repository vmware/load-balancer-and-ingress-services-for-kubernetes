// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPVIPRoutes g c p v IP routes
// swagger:model GCPVIPRoutes
type GCPVIPRoutes struct {

	// Match SE group subnets for VIP placement. Default is to not match SE group subnets. Field introduced in 18.2.9, 20.1.1.
	MatchSeGroupSubnet *bool `json:"match_se_group_subnet,omitempty"`
}

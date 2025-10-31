// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BgpRoutingOptions bgp routing options
// swagger:model BgpRoutingOptions
type BgpRoutingOptions struct {

	// Advertise self as default router to the BGP peer. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdvertiseDefaultRoute *bool `json:"advertise_default_route,omitempty"`

	// Advertise the learned routes to the BGP peer. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdvertiseLearnedRoutes *bool `json:"advertise_learned_routes,omitempty"`

	// Features are applied to peers matching this label. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Label *string `json:"label"`

	// Learn only default route from the BGP peer. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LearnOnlyDefaultRoute *bool `json:"learn_only_default_route,omitempty"`

	// Learn routes from the BGP peer. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LearnRoutes *bool `json:"learn_routes,omitempty"`

	// Maximum number of routes that can be learned from a BGP peer. Allowed values are 50-250. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxLearnLimit *uint32 `json:"max_learn_limit,omitempty"`
}

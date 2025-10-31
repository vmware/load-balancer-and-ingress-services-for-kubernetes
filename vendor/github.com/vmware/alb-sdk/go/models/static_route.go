// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// StaticRoute static route
// swagger:model StaticRoute
type StaticRoute struct {

	// Disable the gateway monitor for default gateway. They are monitored by default. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableGatewayMonitor *bool `json:"disable_gateway_monitor,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IfName *string `json:"if_name,omitempty"`

	// Labels associated with this route. Field introduced in 20.1.1. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Labels []*KeyValue `json:"labels,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NextHop *IPAddr `json:"next_hop"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RouteID *string `json:"route_id"`
}

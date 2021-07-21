// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServicePoolSelector service pool selector
// swagger:model ServicePoolSelector
type ServicePoolSelector struct {

	//  It is a reference to an object of type PoolGroup.
	ServicePoolGroupRef *string `json:"service_pool_group_ref,omitempty"`

	//  It is a reference to an object of type Pool.
	ServicePoolRef *string `json:"service_pool_ref,omitempty"`

	// Pool based destination port. Allowed values are 1-65535.
	// Required: true
	ServicePort *int32 `json:"service_port"`

	// The end of the Service port number range. Allowed values are 1-65535. Special values are 0- 'single port'. Field introduced in 17.2.4.
	ServicePortRangeEnd *int32 `json:"service_port_range_end,omitempty"`

	// Destination protocol to match for the pool selection. If not specified, it will match any protocol. Enum options - PROTOCOL_TYPE_TCP_PROXY, PROTOCOL_TYPE_TCP_FAST_PATH, PROTOCOL_TYPE_UDP_FAST_PATH, PROTOCOL_TYPE_UDP_PROXY.
	ServiceProtocol *string `json:"service_protocol,omitempty"`
}

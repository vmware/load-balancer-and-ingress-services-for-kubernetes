// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServicePoolSelector service pool selector
// swagger:model ServicePoolSelector
type ServicePoolSelector struct {

	//  It is a reference to an object of type PoolGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServicePoolGroupRef *string `json:"service_pool_group_ref,omitempty"`

	//  It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServicePoolRef *string `json:"service_pool_ref,omitempty"`

	// Pool based destination port. Allowed values are 1-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServicePort *uint32 `json:"service_port"`

	// The end of the Service port number range. Allowed values are 1-65535. Special values are 0- single port. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServicePortRangeEnd uint32 `json:"service_port_range_end,omitempty"`

	// Destination protocol to match for the pool selection. If not specified, it will match any protocol. Enum options - PROTOCOL_TYPE_TCP_PROXY, PROTOCOL_TYPE_TCP_FAST_PATH, PROTOCOL_TYPE_UDP_FAST_PATH, PROTOCOL_TYPE_UDP_PROXY, PROTOCOL_TYPE_SCTP_PROXY, PROTOCOL_TYPE_SCTP_FAST_PATH. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceProtocol *string `json:"service_protocol,omitempty"`
}

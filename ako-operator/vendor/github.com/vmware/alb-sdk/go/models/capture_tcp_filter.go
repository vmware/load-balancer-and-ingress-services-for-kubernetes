// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CaptureTCPFilter capture TCP filter
// swagger:model CaptureTCPFilter
type CaptureTCPFilter struct {

	// Destination Port range filter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DstPortRange *DestinationPortAddr `json:"dst_port_range,omitempty"`

	// Ethernet Proto filter. Enum options - ETH_TYPE_IPV4. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EthProto *string `json:"eth_proto,omitempty"`

	// Per packet IP filter for Service Engine PCAP. Matches with source and destination address. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HostIP *DebugIPAddr `json:"host_ip,omitempty"`

	// Source Port range filter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SrcPortRange *SourcePortAddr `json:"src_port_range,omitempty"`

	// TCP flags filter. Or'ed internally and And'ed amongst each other. . Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Tcpflag *CaptureTCPFlags `json:"tcpflag,omitempty"`
}

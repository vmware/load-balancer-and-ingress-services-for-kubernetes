// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkProfileUnion network profile union
// swagger:model NetworkProfileUnion
type NetworkProfileUnion struct {

	// Configure SCTP FastPath network profile. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SctpFastPathProfile *SCTPFastPathProfile `json:"sctp_fast_path_profile,omitempty"`

	// Configure SCTP Proxy network profile. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SctpProxyProfile *SCTPProxyProfile `json:"sctp_proxy_profile,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPFastPathProfile *TCPFastPathProfile `json:"tcp_fast_path_profile,omitempty"`

	//  Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	TCPProxyProfile *TCPProxyProfile `json:"tcp_proxy_profile,omitempty"`

	// Configure one of either proxy or fast path profiles. Enum options - PROTOCOL_TYPE_TCP_PROXY, PROTOCOL_TYPE_TCP_FAST_PATH, PROTOCOL_TYPE_UDP_FAST_PATH, PROTOCOL_TYPE_UDP_PROXY, PROTOCOL_TYPE_SCTP_PROXY, PROTOCOL_TYPE_SCTP_FAST_PATH. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- PROTOCOL_TYPE_TCP_FAST_PATH,PROTOCOL_TYPE_UDP_FAST_PATH), Basic edition(Allowed values- PROTOCOL_TYPE_TCP_PROXY,PROTOCOL_TYPE_TCP_FAST_PATH,PROTOCOL_TYPE_UDP_FAST_PATH), Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UDPFastPathProfile *UDPFastPathProfile `json:"udp_fast_path_profile,omitempty"`

	// Configure UDP Proxy network profile. Field introduced in 17.2.8, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UDPProxyProfile *UDPProxyProfile `json:"udp_proxy_profile,omitempty"`
}

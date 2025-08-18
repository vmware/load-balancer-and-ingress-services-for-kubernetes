// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeRateLimiters se rate limiters
// swagger:model SeRateLimiters
type SeRateLimiters struct {

	// Rate limiter for ARP packets in pps. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ArpRl *uint32 `json:"arp_rl,omitempty"`

	// Default Rate limiter in pps. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DefaultRl *uint32 `json:"default_rl,omitempty"`

	// Rate limiter for number of flow probes in pps. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowProbeRl *uint32 `json:"flow_probe_rl,omitempty"`

	// Rate limiter for ICMP requests in pps. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcmpRl *uint32 `json:"icmp_rl,omitempty"`

	// Rate limiter for ICMP response in pps. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcmpRspRl *uint32 `json:"icmp_rsp_rl,omitempty"`

	// Rate limiter for number RST pkts sent in pps. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RstRl *uint32 `json:"rst_rl,omitempty"`
}

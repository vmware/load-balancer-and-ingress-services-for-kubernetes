// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IptableRule iptable rule
// swagger:model IptableRule
type IptableRule struct {

	//  Enum options - ACCEPT, DROP, REJECT, DNAT, MASQUERADE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Action *string `json:"action"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DnatIP *IPAddr `json:"dnat_ip,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DstIP *IPAddrPrefix `json:"dst_ip,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DstPort *PortRange `json:"dst_port,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InputInterface *string `json:"input_interface,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OutputInterface *string `json:"output_interface,omitempty"`

	//  Enum options - PROTO_TCP, PROTO_UDP, PROTO_ICMP, PROTO_ALL. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Proto *string `json:"proto,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SrcIP *IPAddrPrefix `json:"src_ip,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SrcPort *PortRange `json:"src_port,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tag *string `json:"tag,omitempty"`
}

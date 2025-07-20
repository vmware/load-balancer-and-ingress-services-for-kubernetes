// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSEdnsOption Dns edns option
// swagger:model DnsEdnsOption
type DNSEdnsOption struct {

	// Address family. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AddrFamily *uint32 `json:"addr_family,omitempty"`

	// EDNS option code. Enum options - EDNS_OPTION_CODE_NSID, EDNS_OPTION_CODE_DNSSEC_DAU, EDNS_OPTION_CODE_DNSSEC_DHU, EDNS_OPTION_CODE_DNSSEC_N3U, EDNS_OPTION_CODE_CLIENT_SUBNET, EDNS_OPTION_CODE_EXPIRE, EDNS_OPTION_CODE_COOKIE, EDNS_OPTION_CODE_TCP_KEEPALIVE, EDNS_OPTION_CODE_PADDING, EDNS_OPTION_CODE_CHAIN. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Code *string `json:"code"`

	// Scope prefix length of address. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScopePrefixLen *uint32 `json:"scope_prefix_len,omitempty"`

	// Source prefix length of address. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SourcePrefixLen *uint32 `json:"source_prefix_len,omitempty"`

	// IPv4 address of the client subnet. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubnetIP *uint32 `json:"subnet_ip,omitempty"`

	// IPv6 address of the client subnet. Field introduced in 18.2.12, 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SubnetIp6 *string `json:"subnet_ip6,omitempty"`
}

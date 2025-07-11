// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSNsRdata Dns ns rdata
// swagger:model DnsNsRdata
type DNSNsRdata struct {

	// IPv6 address for Name Server. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ip6Address *IPAddr `json:"ip6_address,omitempty"`

	// IP address for Name Server. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAddress *IPAddr `json:"ip_address,omitempty"`

	// Name Server name. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Nsname *string `json:"nsname"`
}

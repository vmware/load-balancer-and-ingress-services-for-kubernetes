// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSClientIPMatch Dns client Ip match
// swagger:model DnsClientIpMatch
type DNSClientIPMatch struct {

	// IP addresses to match against client IP. Field introduced in 17.1.6,17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientIP *IPAddrMatch `json:"client_ip"`

	// Use the IP address from the EDNS client subnet option, if available, as the source IP address of the client. It should be noted that the edns subnet IP may not be a /32 IP address. Field introduced in 17.1.6,17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseEdnsClientSubnetIP *bool `json:"use_edns_client_subnet_ip,omitempty"`
}

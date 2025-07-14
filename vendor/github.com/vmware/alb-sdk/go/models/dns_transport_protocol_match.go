// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSTransportProtocolMatch Dns transport protocol match
// swagger:model DnsTransportProtocolMatch
type DNSTransportProtocolMatch struct {

	// Criterion to use for matching the DNS transport protocol. Enum options - IS_IN, IS_NOT_IN. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Protocol to match against transport protocol used by DNS query. Enum options - DNS_OVER_UDP, DNS_OVER_TCP. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Protocol *string `json:"protocol"`
}

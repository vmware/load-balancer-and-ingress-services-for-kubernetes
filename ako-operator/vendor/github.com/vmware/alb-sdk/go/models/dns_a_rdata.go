// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSARdata Dns a rdata
// swagger:model DnsARdata
type DNSARdata struct {

	// IP address for FQDN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IPAddress *IPAddr `json:"ip_address"`
}

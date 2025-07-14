// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSConfiguration DNS configuration
// swagger:model DNSConfiguration
type DNSConfiguration struct {

	// Search domain to use in DNS lookup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SearchDomain *string `json:"search_domain,omitempty"`

	// List of DNS Server IP(v4/v6) addresses or FQDNs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerList []*IPAddr `json:"server_list,omitempty"`
}

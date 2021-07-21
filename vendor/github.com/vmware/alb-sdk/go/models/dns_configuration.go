// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSConfiguration DNS configuration
// swagger:model DNSConfiguration
type DNSConfiguration struct {

	// Search domain to use in DNS lookup.
	SearchDomain *string `json:"search_domain,omitempty"`

	// List of DNS Server IP addresses.
	ServerList []*IPAddr `json:"server_list,omitempty"`
}

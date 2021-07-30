// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSResolver Dns resolver
// swagger:model DnsResolver
type DNSResolver struct {

	// If configured, this value used for refreshing the DNS entries.Overrides both received_ttl and min_ttl. The entries are refreshed only on fixed_ttleven when received_ttl is less than fixed_ttl. Allowed values are 5-2147483647. Field introduced in 20.1.5. Unit is SEC.
	FixedTTL *int32 `json:"fixed_ttl,omitempty"`

	// If configured, this ttl overrides the ttl from responses if ttl < min_ttl.effectively ttl = max(recieved_ttl, min_ttl). Allowed values are 5-2147483647. Field introduced in 20.1.5. Unit is SEC.
	MinTTL *int32 `json:"min_ttl,omitempty"`

	// Name server IPv4 addresses. Field introduced in 20.1.5. Minimum of 1 items required. Maximum of 10 items allowed.
	NameserverIps []*IPAddr `json:"nameserver_ips,omitempty"`

	// Unique name for resolver config. Field introduced in 20.1.5.
	// Required: true
	ResolverName *string `json:"resolver_name"`

	// If enabled, DNS resolution is performed via management network. Field introduced in 20.1.5.
	UseMgmt *bool `json:"use_mgmt,omitempty"`
}

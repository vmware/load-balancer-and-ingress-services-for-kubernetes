// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSResolver Dns resolver
// swagger:model DnsResolver
type DNSResolver struct {

	// If configured, this value used for refreshing the DNS entries.Overrides both received_ttl and min_ttl. The entries are refreshed only on fixed_ttleven when received_ttl is less than fixed_ttl. Allowed values are 5-2147483647. Field introduced in 20.1.5. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FixedTTL *uint32 `json:"fixed_ttl,omitempty"`

	// If configured, the min_ttl overrides the ttl from responses when ttl < min_ttl,effectively ttl = max(recieved_ttl, min_ttl). Allowed values are 5-2147483647. Field introduced in 20.1.5. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinTTL *uint32 `json:"min_ttl,omitempty"`

	// Name server IPv4 addresses. Field introduced in 20.1.5. Minimum of 1 items required. Maximum of 10 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NameserverIps []*IPAddr `json:"nameserver_ips,omitempty"`

	// Unique name for resolver config. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ResolverName *string `json:"resolver_name"`

	// If enabled, DNS resolution is performed via management network. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UseMgmt *bool `json:"use_mgmt,omitempty"`
}

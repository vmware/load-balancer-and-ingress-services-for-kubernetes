// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAddrGroup Ip addr group
// swagger:model IpAddrGroup
type IPAddrGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Configure IP address(es). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Addrs []*IPAddr `json:"addrs,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Populate the IP address ranges from the geo database for this country. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CountryCodes []string `json:"country_codes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Configure (IP address, port) tuple(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPPorts []*IPAddrPort `json:"ip_ports,omitempty"`

	// Populate IP addresses from tasks of this Marathon app. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonAppName *string `json:"marathon_app_name,omitempty"`

	// Task port associated with marathon service port. If Marathon app has multiple service ports, this is required. Else, the first task port is used. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonServicePort *uint32 `json:"marathon_service_port,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Name of the IP address group. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Configure IP address prefix(es). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Prefixes []*IPAddrPrefix `json:"prefixes,omitempty"`

	// Configure IP address range(s). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ranges []*IPAddrRange `json:"ranges,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the IP address group. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

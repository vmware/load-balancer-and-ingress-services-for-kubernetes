// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TenantSystemConfiguration tenant system configuration
// swagger:model TenantSystemConfiguration
type TenantSystemConfiguration struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// DNS virtual services hosting FQDN records for applications configured within this tenant. It is a reference to an object of type VirtualService. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSVirtualserviceRefs []string `json:"dns_virtualservice_refs,omitempty"`

	// Name of the Tenant System Configuration object. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Unique Identifier of the tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Tenant System Configuration UUID. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

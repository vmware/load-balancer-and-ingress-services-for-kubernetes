// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotIPReputationTypeMapping bot IP reputation type mapping
// swagger:model BotIPReputationTypeMapping
type BotIPReputationTypeMapping struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Map every IPReputationType to a bot type (can be unknown). Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPReputationMappings []*IPReputationTypeMapping `json:"ip_reputation_mappings,omitempty"`

	// The name of this mapping. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// The unique identifier of the tenant to which this mapping belongs. It is a reference to an object of type Tenant. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// A unique identifier of this mapping. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

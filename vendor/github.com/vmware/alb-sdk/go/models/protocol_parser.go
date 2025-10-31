// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ProtocolParser protocol parser
// swagger:model ProtocolParser
type ProtocolParser struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Description of the protocol parser. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Name of the protocol parser. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Command script provided inline. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ParserCode *string `json:"parser_code"`

	// Tenant UUID of the protocol parser. It is a reference to an object of type Tenant. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the protocol parser. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

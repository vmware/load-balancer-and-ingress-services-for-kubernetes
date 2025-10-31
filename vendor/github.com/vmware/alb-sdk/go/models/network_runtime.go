// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NetworkRuntime network runtime
// swagger:model NetworkRuntime
type NetworkRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Objects using static IPs in this network. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjUuids []string `json:"obj_uuids,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubnetRuntime []*SubnetRuntime `json:"subnet_runtime,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

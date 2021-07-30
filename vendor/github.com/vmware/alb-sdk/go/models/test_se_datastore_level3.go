// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TestSeDatastoreLevel3 test se datastore level3
// swagger:model TestSeDatastoreLevel3
type TestSeDatastoreLevel3 struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 18.2.6.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Role role
// swagger:model Role
type Role struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allow access to unlabelled objects. Field introduced in 20.1.5.
	AllowUnlabelledAccess *bool `json:"allow_unlabelled_access,omitempty"`

	// Filters for granular object access control based on object labels. Multiple filters are merged using the AND operator. If empty, all objects according to the privileges will be accessible to the user. Field introduced in 20.1.3. Maximum of 4 items allowed.
	Filters []*RoleFilter `json:"filters,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property privileges of obj type Role field type str  type object
	Privileges []*Permission `json:"privileges,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}

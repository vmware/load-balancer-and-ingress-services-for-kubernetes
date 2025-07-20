// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UserRole user role
// swagger:model UserRole
type UserRole struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllTenants *bool `json:"all_tenants,omitempty"`

	//  It is a reference to an object of type Role. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RoleRef *string `json:"role_ref,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`
}

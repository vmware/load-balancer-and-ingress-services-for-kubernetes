// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpenStackRoleMapping open stack role mapping
// swagger:model OpenStackRoleMapping
type OpenStackRoleMapping struct {

	// Role name in Avi. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AviRole *string `json:"avi_role"`

	// Role name in OpenStack. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	OsRole *string `json:"os_role"`
}

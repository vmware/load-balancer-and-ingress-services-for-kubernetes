// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpenStackClusterSetup open stack cluster setup
// swagger:model OpenStackClusterSetup
type OpenStackClusterSetup struct {

	// admin_tenant of OpenStackClusterSetup.
	AdminTenant *string `json:"admin_tenant,omitempty"`

	// auth_url of OpenStackClusterSetup.
	AuthURL *string `json:"auth_url,omitempty"`

	// cc_id of OpenStackClusterSetup.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of OpenStackClusterSetup.
	ErrorString *string `json:"error_string,omitempty"`

	// keystone_host of OpenStackClusterSetup.
	// Required: true
	KeystoneHost *string `json:"keystone_host"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`
}

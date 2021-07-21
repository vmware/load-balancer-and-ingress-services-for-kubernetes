// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIAdminCredentials v i admin credentials
// swagger:model VIAdminCredentials
type VIAdminCredentials struct {

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// password of VIAdminCredentials.
	Password *string `json:"password,omitempty"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`

	// vcenter_url of VIAdminCredentials.
	// Required: true
	VcenterURL *string `json:"vcenter_url"`

	// vi_mgr_token of VIAdminCredentials.
	ViMgrToken *string `json:"vi_mgr_token,omitempty"`
}

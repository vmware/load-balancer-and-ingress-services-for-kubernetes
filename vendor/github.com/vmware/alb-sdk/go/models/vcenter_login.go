// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterLogin vcenter login
// swagger:model VcenterLogin
type VcenterLogin struct {

	// Unique object identifier of cloud.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	// password of VcenterLogin.
	Password *string `json:"password,omitempty"`

	// Number of start_ts.
	StartTs *int64 `json:"start_ts,omitempty"`

	// username of VcenterLogin.
	Username *string `json:"username,omitempty"`

	// vcenter_url of VcenterLogin.
	VcenterURL *string `json:"vcenter_url,omitempty"`
}

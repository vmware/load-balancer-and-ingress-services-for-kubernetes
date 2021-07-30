// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpenstackLogin openstack login
// swagger:model OpenstackLogin
type OpenstackLogin struct {

	// admin_tenant of OpenstackLogin.
	AdminTenant *string `json:"admin_tenant,omitempty"`

	// auth_url of OpenstackLogin.
	AuthURL *string `json:"auth_url,omitempty"`

	// keystone_host of OpenstackLogin.
	KeystoneHost *string `json:"keystone_host,omitempty"`

	// password of OpenstackLogin.
	// Required: true
	Password *string `json:"password"`

	// region of OpenstackLogin.
	Region *string `json:"region,omitempty"`

	// username of OpenstackLogin.
	// Required: true
	Username *string `json:"username"`
}

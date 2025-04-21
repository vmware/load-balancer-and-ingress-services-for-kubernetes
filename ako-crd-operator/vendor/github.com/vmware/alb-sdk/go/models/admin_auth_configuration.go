// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AdminAuthConfiguration admin auth configuration
// swagger:model AdminAuthConfiguration
type AdminAuthConfiguration struct {

	// Allow any user created locally to login with local credentials. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowLocalUserLogin *bool `json:"allow_local_user_login,omitempty"`

	// Remote Auth configurations. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	RemoteAuthConfigurations []*RemoteAuthConfiguration `json:"remote_auth_configurations,omitempty"`

	// Service Auth configurations. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceAuthConfigurations []*ServiceAuthConfiguration `json:"service_auth_configurations,omitempty"`
}

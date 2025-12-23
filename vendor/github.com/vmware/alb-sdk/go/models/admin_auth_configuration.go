// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AdminAuthConfiguration admin auth configuration
// swagger:model AdminAuthConfiguration
type AdminAuthConfiguration struct {

	// Allow any user created locally to login with local credentials. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowLocalUserLogin *bool `json:"allow_local_user_login,omitempty"`

	// Secondary authentication mechanisms to be used. Field deprecated in 22.1.1. Field introduced in 20.1.6. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AlternateAuthConfigurations []*AlternateAuthConfiguration `json:"alternate_auth_configurations,omitempty"`

	//  It is a reference to an object of type AuthProfile. Field deprecated in 22.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthProfileRef *string `json:"auth_profile_ref,omitempty"`

	// Rules list for tenant or role mapping. Field deprecated in 22.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MappingRules []*AuthMappingRule `json:"mapping_rules,omitempty"`

	// Remote Auth configurations. Field introduced in 22.1.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	RemoteAuthConfigurations []*RemoteAuthConfiguration `json:"remote_auth_configurations,omitempty"`
}

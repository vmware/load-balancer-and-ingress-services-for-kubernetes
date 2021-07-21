// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AdminAuthConfiguration admin auth configuration
// swagger:model AdminAuthConfiguration
type AdminAuthConfiguration struct {

	// Allow any user created locally to login with local credentials. Field introduced in 17.1.1.
	AllowLocalUserLogin *bool `json:"allow_local_user_login,omitempty"`

	// Secondary authentication mechanisms to be used. Field introduced in 20.1.6. Maximum of 1 items allowed.
	AlternateAuthConfigurations []*AlternateAuthConfiguration `json:"alternate_auth_configurations,omitempty"`

	//  It is a reference to an object of type AuthProfile.
	AuthProfileRef *string `json:"auth_profile_ref,omitempty"`

	// Rules list for tenant or role mapping.
	MappingRules []*AuthMappingRule `json:"mapping_rules,omitempty"`
}

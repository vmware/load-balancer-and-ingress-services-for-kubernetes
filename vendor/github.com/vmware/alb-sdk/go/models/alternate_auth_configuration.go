// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlternateAuthConfiguration alternate auth configuration
// swagger:model AlternateAuthConfiguration
type AlternateAuthConfiguration struct {

	// UUID of the authprofile. It is a reference to an object of type AuthProfile. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AuthProfileRef *string `json:"auth_profile_ref,omitempty"`

	// index used for maintaining order of AlternateAuthConfiguration. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	// Rules list for tenant or role mapping. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MappingRules []*AuthMappingRule `json:"mapping_rules,omitempty"`
}

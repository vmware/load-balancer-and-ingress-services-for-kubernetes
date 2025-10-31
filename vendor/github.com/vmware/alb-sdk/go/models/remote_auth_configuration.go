// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RemoteAuthConfiguration remote auth configuration
// swagger:model RemoteAuthConfiguration
type RemoteAuthConfiguration struct {

	// UUID of the AuthMappingProfile(set of auth mapping rules) to be assigned to a user on successful match. It is a reference to an object of type AuthMappingProfile. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	AuthMappingProfileRef *string `json:"auth_mapping_profile_ref"`

	// UUID of the auth profile. It is a reference to an object of type AuthProfile. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	AuthProfileRef *string `json:"auth_profile_ref"`

	// Index used for maintaining order of RemoteAuthConfiguration. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`
}

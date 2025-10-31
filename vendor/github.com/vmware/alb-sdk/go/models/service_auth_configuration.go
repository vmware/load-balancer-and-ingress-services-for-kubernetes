// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceAuthConfiguration service auth configuration
// swagger:model ServiceAuthConfiguration
type ServiceAuthConfiguration struct {

	// Index used for maintaining order of ServiceAuthConfiguration. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	// UUID of the AuthMappingProfile(set of auth mapping rules) to be assigned to a user on successful match. It is a reference to an object of type AuthMappingProfile. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ServiceAuthMappingProfileRef *string `json:"service_auth_mapping_profile_ref"`

	// UUID of the service auth profile. It is a reference to an object of type ServiceAuthProfile. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ServiceAuthProfileRef *string `json:"service_auth_profile_ref"`
}

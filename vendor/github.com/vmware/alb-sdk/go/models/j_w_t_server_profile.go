// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JWTServerProfile j w t server profile
// swagger:model JWTServerProfile
type JWTServerProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// JWT Auth configuration for profile_type CONTROLLER_INTERNAL_AUTH. Field introduced in 20.1.6.
	ControllerInternalAuth *ControllerInternalAuth `json:"controller_internal_auth,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster.  If the field is set to true, then the object is replicated across the federation.  . Field introduced in 20.1.6.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Uniquely identifiable name of the Token Issuer, only allowed with profile_type CLIENT_AUTH. Field introduced in 20.1.3.
	Issuer *string `json:"issuer,omitempty"`

	// JWKS key set used for validating the JWT, only allowed with profile_type CLIENT_AUTH. Field introduced in 20.1.3.
	JwksKeys *string `json:"jwks_keys,omitempty"`

	// Type of JWT Server profile which defines the usage type. Enum options - CLIENT_AUTH, CONTROLLER_INTERNAL_AUTH. Field introduced in 20.1.6.
	JwtProfileType *string `json:"jwt_profile_type,omitempty"`

	// Name of the JWT Profile. Field introduced in 20.1.3.
	// Required: true
	Name *string `json:"name"`

	// UUID of the Tenant. It is a reference to an object of type Tenant. Field introduced in 20.1.3.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the JWTProfile. Field introduced in 20.1.3.
	UUID *string `json:"uuid,omitempty"`
}

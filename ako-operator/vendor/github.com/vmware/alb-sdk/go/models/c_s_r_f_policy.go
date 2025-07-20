// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CSRFPolicy c s r f policy
// swagger:model CSRFPolicy
type CSRFPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Name of the cookie to be used for CSRF token. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CookieName *string `json:"cookie_name,omitempty"`

	// Human-readable description of this CSRF Protection Policy. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// The name of this CSRF Protection Policy. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Rules to control which requests undergo CSRF Protection.If the client's request doesn't match with any rules MatchTarget, BYPASS_CSRF action is applied. Field introduced in 30.2.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Rules []*CSRFRule `json:"rules,omitempty"`

	// The unique identifier of the tenant to which this policy belongs. It is a reference to an object of type Tenant. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// CSRF token is rotated when this time expires. Tokens will be acceptable for twice the token_validity_time time. Allowed values are 10-1440. Special values are 0- unlimited. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenValidityTimeMin *uint32 `json:"token_validity_time_min,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// A unique identifier to this CSRF Protection Policy. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

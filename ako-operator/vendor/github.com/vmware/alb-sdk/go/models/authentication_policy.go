// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthenticationPolicy authentication policy
// swagger:model AuthenticationPolicy
type AuthenticationPolicy struct {

	// Add rules to apply auth profile to specific targets. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthnRules []*AuthenticationRule `json:"authn_rules,omitempty"`

	// Auth Profile to use for validating users. It is a reference to an object of type AuthProfile. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DefaultAuthProfileRef *string `json:"default_auth_profile_ref,omitempty"`
}

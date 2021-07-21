// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthorizationPolicy authorization policy
// swagger:model AuthorizationPolicy
type AuthorizationPolicy struct {

	// Authorization Policy Rules. Field introduced in 18.2.5.
	AuthzRules []*AuthorizationRule `json:"authz_rules,omitempty"`
}

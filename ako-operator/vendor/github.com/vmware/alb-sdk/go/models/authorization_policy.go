// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthorizationPolicy authorization policy
// swagger:model AuthorizationPolicy
type AuthorizationPolicy struct {

	// Authorization Policy Rules. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthzRules []*AuthorizationRule `json:"authz_rules,omitempty"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicyAllowlist waf policy allowlist
// swagger:model WafPolicyAllowlist
type WafPolicyAllowlist struct {

	// Rules to bypass WAF. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Rules []*WafPolicyAllowlistRule `json:"rules,omitempty"`
}

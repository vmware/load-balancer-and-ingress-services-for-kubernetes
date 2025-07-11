// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafAllowlistLog waf allowlist log
// swagger:model WafAllowlistLog
type WafAllowlistLog struct {

	// Actions generated by this rule. Enum options - WAF_POLICY_ALLOWLIST_ACTION_BYPASS, WAF_POLICY_ALLOWLIST_ACTION_DETECTION_MODE, WAF_POLICY_ALLOWLIST_ACTION_CONTINUE. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Actions []string `json:"actions,omitempty"`

	// Name of the matched rule. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleName *string `json:"rule_name,omitempty"`
}

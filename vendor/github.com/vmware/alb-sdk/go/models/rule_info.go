// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RuleInfo rule info
// swagger:model RuleInfo
type RuleInfo struct {

	// URI hitted signature rule matches. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Matches []*Matches `json:"matches,omitempty"`

	// URI hitted signature rule group id. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleGroupID *string `json:"rule_group_id,omitempty"`

	// URI hitted signature rule id. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleID *string `json:"rule_id,omitempty"`
}

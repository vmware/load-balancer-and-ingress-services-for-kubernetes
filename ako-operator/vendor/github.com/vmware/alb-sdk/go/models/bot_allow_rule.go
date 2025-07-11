// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotAllowRule bot allow rule
// swagger:model BotAllowRule
type BotAllowRule struct {

	// The action to take. Enum options - BOT_ACTION_BYPASS, BOT_ACTION_CONTINUE. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Action *string `json:"action"`

	// The condition to match. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Condition *MatchTarget `json:"condition"`

	// Rules are processed in order of this index field. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	// A name describing the rule in a short form. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`
}

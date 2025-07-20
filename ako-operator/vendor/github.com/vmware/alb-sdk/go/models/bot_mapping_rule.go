// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotMappingRule bot mapping rule
// swagger:model BotMappingRule
type BotMappingRule struct {

	// The assigned classification for this client. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Classification *BotClassification `json:"classification"`

	// Rules are processed in order of this index field. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	// How to match the request  all the specified properties must be fulfilled. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Match *BotMappingRuleMatchTarget `json:"match"`

	// A name describing the rule in a short form. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`
}

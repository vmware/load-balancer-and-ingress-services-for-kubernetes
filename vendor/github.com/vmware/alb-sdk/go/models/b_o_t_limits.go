// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BOTLimits b o t limits
// swagger:model BOTLimits
type BOTLimits struct {

	// Maximum number of rules to control which requests undergo BOT detection. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AllowRules *int32 `json:"allow_rules,omitempty"`

	// Maximum number of configurable HTTP header(s). Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Hdrs *int32 `json:"hdrs,omitempty"`

	// Maximum number of rules in a BotMapping object. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MappingRules *int32 `json:"mapping_rules,omitempty"`
}

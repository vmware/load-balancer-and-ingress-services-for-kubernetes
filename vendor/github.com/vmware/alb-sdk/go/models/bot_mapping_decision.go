// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotMappingDecision bot mapping decision
// swagger:model BotMappingDecision
type BotMappingDecision struct {

	// The name of the Bot Mapping that made the decision. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MappingName *string `json:"mapping_name,omitempty"`

	// The name of the Bot Mapping Rule that made the decision. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MappingRuleName *string `json:"mapping_rule_name,omitempty"`
}

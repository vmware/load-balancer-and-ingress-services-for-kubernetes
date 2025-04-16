// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotManagementLog bot management log
// swagger:model BotManagementLog
type BotManagementLog struct {

	// The final classification of the bot management module. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Classification *BotClassification `json:"classification,omitempty"`

	// Bot Mapping details. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MappingDecision *BotMappingDecision `json:"mapping_decision,omitempty"`

	// The evaluation results of the various bot module components. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Results []*BotEvaluationResult `json:"results,omitempty"`
}

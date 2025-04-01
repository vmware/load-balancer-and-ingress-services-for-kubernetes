// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotEvaluationResult bot evaluation result
// swagger:model BotEvaluationResult
type BotEvaluationResult struct {

	// The component of the bot module that made this evaluation. Enum options - BOT_DECIDER_CONSOLIDATION, BOT_DECIDER_USER_AGENT, BOT_DECIDER_IP_REPUTATION, BOT_DECIDER_IP_NETWORK_LOCATION, BOT_DECIDER_CLIENT_BEHAVIOR. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Component *string `json:"component,omitempty"`

	// The confidence of this evaluation. Enum options - LOW_CONFIDENCE, MEDIUM_CONFIDENCE, HIGH_CONFIDENCE. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Confidence *string `json:"confidence,omitempty"`

	// The resulting Bot Identification. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Identification *BotIdentification `json:"identification,omitempty"`

	// Additional notes for this result. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Notes []string `json:"notes,omitempty"`
}

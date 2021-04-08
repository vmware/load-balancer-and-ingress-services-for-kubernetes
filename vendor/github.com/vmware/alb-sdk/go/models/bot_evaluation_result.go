package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotEvaluationResult bot evaluation result
// swagger:model BotEvaluationResult
type BotEvaluationResult struct {

	// The component of the bot module that made this evaluation. Enum options - BOT_DECIDER_CONSOLIDATION, BOT_DECIDER_USER_AGENT, BOT_DECIDER_IP_REPUTATION, BOT_DECIDER_IP_NETWORK_LOCATION. Field introduced in 21.1.1.
	Component *string `json:"component,omitempty"`

	// The confidence of this evaluation. Enum options - LOW_CONFIDENCE, MEDIUM_CONFIDENCE, HIGH_CONFIDENCE. Field introduced in 21.1.1.
	Confidence *string `json:"confidence,omitempty"`

	// The resultion Bot Identification. Field introduced in 21.1.1.
	Identification *BotIdentification `json:"identification,omitempty"`

	// Additional notes for this result. Field introduced in 21.1.1.
	Notes []string `json:"notes,omitempty"`
}

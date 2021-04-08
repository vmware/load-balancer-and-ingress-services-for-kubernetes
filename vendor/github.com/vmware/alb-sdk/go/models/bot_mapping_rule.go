package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotMappingRule bot mapping rule
// swagger:model BotMappingRule
type BotMappingRule struct {

	// How to match the BotClientClass. Field introduced in 21.1.1.
	ClassMatcher *BotClassMatcher `json:"class_matcher,omitempty"`

	// The assigned classification for this client. Field introduced in 21.1.1.
	// Required: true
	Classification *BotClassification `json:"classification"`

	// The component for which this mapping is used. Enum options - BOT_DECIDER_CONSOLIDATION, BOT_DECIDER_USER_AGENT, BOT_DECIDER_IP_REPUTATION, BOT_DECIDER_IP_NETWORK_LOCATION. Field introduced in 21.1.1.
	ComponentMatcher *string `json:"component_matcher,omitempty"`

	// The list of bot identifier names and how they're matched. Field introduced in 21.1.1.
	IdentifierMatcher *StringMatch `json:"identifier_matcher,omitempty"`

	// How to match the BotClientType. Field introduced in 21.1.1.
	TypeMatcher *BotTypeMatcher `json:"type_matcher,omitempty"`
}

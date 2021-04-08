package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotClassMatcher bot class matcher
// swagger:model BotClassMatcher
type BotClassMatcher struct {

	// The list of client classes. Enum options - HUMAN_CLIENT, BOT_CLIENT. Field introduced in 21.1.1.
	ClientClasses []string `json:"client_classes,omitempty"`

	// The match operation. Enum options - IS_IN, IS_NOT_IN. Field introduced in 21.1.1.
	Op *string `json:"op,omitempty"`
}

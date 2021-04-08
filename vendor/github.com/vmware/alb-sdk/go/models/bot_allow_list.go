package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotAllowList bot allow list
// swagger:model BotAllowList
type BotAllowList struct {

	// Allow rules to control which requests undergo BOT detection. Field introduced in 21.1.1.
	Rules []*BotAllowRule `json:"rules,omitempty"`
}

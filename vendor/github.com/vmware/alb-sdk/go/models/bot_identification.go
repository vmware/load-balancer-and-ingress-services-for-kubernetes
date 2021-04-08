package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotIdentification bot identification
// swagger:model BotIdentification
type BotIdentification struct {

	// The Bot Client Class of this identification. Enum options - HUMAN_CLIENT, BOT_CLIENT. Field introduced in 21.1.1.
	// Required: true
	Class *string `json:"class"`

	// A free-form *string to identify the client. Field introduced in 21.1.1.
	// Required: true
	Identifier *string `json:"identifier"`

	// The Bot Client Type of this identification. Enum options - WEB_BROWSER, IN_APP_BROWSER, SEARCH_ENGINE, IMPERSONATOR, SPAM_SOURCE, WEB_ATTACKS, BOTNET, SCANNER, DENIAL_OF_SERVICE, CLOUD_SOURCE. Field introduced in 21.1.1.
	Type *string `json:"type,omitempty"`
}

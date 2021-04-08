package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotTypeMatcher bot type matcher
// swagger:model BotTypeMatcher
type BotTypeMatcher struct {

	// The list of client types. Enum options - WEB_BROWSER, IN_APP_BROWSER, SEARCH_ENGINE, IMPERSONATOR, SPAM_SOURCE, WEB_ATTACKS, BOTNET, SCANNER, DENIAL_OF_SERVICE, CLOUD_SOURCE. Field introduced in 21.1.1.
	ClientTypes []string `json:"client_types,omitempty"`

	// The match operation. Enum options - IS_IN, IS_NOT_IN. Field introduced in 21.1.1.
	Op *string `json:"op,omitempty"`
}

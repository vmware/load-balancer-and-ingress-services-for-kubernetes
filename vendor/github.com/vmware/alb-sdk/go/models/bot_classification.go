package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotClassification bot classification
// swagger:model BotClassification
type BotClassification struct {

	// One of the system-defined Bot classification types. Enum options - HUMAN, GOOD_BOT, BAD_BOT, DANGEROUS_BOT, USER_DEFINED_BOT, UNKNOWN_CLIENT. Field introduced in 21.1.1.
	// Required: true
	Type *string `json:"type"`

	// If 'type' has BotClassificationTypes value 'USER_DEFINED', this is the user-defined value. Field introduced in 21.1.1.
	UserDefinedType *string `json:"user_defined_type,omitempty"`
}

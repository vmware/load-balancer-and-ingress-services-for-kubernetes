package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthenticationAction authentication action
// swagger:model AuthenticationAction
type AuthenticationAction struct {

	// Authentication Action to be taken for a matched Rule. Enum options - SKIP_AUTHENTICATION, USE_DEFAULT_AUTHENTICATION. Field introduced in 18.2.5.
	// Required: true
	Type *string `json:"type"`
}

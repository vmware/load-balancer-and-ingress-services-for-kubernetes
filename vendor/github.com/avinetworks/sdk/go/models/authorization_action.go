package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthorizationAction authorization action
// swagger:model AuthorizationAction
type AuthorizationAction struct {

	// HTTP status code to use for local response when an policy rule is matched. Enum options - HTTP_RESPONSE_STATUS_CODE_403. Field introduced in 18.2.5.
	StatusCode *string `json:"status_code,omitempty"`

	// Defines the action taken when an authorization policy rule is matched.By default, access is allowed to the requested resource. Enum options - ALLOW_ACCESS, CLOSE_CONNECTION, HTTP_LOCAL_RESPONSE. Field introduced in 18.2.5.
	Type *string `json:"type,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RateLimiterAction rate limiter action
// swagger:model RateLimiterAction
type RateLimiterAction struct {

	// File to be used for HTTP Local response rate limit action.
	File *HTTPLocalFile `json:"file,omitempty"`

	// Parameters for HTTP Redirect rate limit action.
	Redirect *HTTPRedirectAction `json:"redirect,omitempty"`

	// HTTP status code for Local Response rate limit action. Enum options - HTTP_LOCAL_RESPONSE_STATUS_CODE_200, HTTP_LOCAL_RESPONSE_STATUS_CODE_204, HTTP_LOCAL_RESPONSE_STATUS_CODE_403, HTTP_LOCAL_RESPONSE_STATUS_CODE_404, HTTP_LOCAL_RESPONSE_STATUS_CODE_429, HTTP_LOCAL_RESPONSE_STATUS_CODE_501.
	StatusCode *string `json:"status_code,omitempty"`

	// Type of action to be enforced upon hitting the rate limit. Enum options - RL_ACTION_NONE, RL_ACTION_DROP_CONN, RL_ACTION_RESET_CONN, RL_ACTION_CLOSE_CONN, RL_ACTION_LOCAL_RSP, RL_ACTION_REDIRECT.
	Type *string `json:"type,omitempty"`
}

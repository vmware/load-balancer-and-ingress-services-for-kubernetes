package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HttpsecurityAction httpsecurity action
// swagger:model HTTPSecurityAction
type HttpsecurityAction struct {

	// Type of the security action to perform. Enum options - HTTP_SECURITY_ACTION_CLOSE_CONN, HTTP_SECURITY_ACTION_SEND_RESPONSE, HTTP_SECURITY_ACTION_ALLOW, HTTP_SECURITY_ACTION_REDIRECT_TO_HTTPS, HTTP_SECURITY_ACTION_RATE_LIMIT.
	// Required: true
	Action *string `json:"action"`

	// File to be used for generating HTTP local response.
	File *HTTPLocalFile `json:"file,omitempty"`

	// Secure SSL/TLS port to redirect the HTTP request to. Allowed values are 1-65535.
	HTTPSPort *int32 `json:"https_port,omitempty"`

	// Rate Limit profile to be used to rate-limit the flow.
	RateLimit *RateProfile `json:"rate_limit,omitempty"`

	// HTTP status code to use for local response. Enum options - HTTP_LOCAL_RESPONSE_STATUS_CODE_200, HTTP_LOCAL_RESPONSE_STATUS_CODE_204, HTTP_LOCAL_RESPONSE_STATUS_CODE_403, HTTP_LOCAL_RESPONSE_STATUS_CODE_404, HTTP_LOCAL_RESPONSE_STATUS_CODE_429, HTTP_LOCAL_RESPONSE_STATUS_CODE_501.
	StatusCode *string `json:"status_code,omitempty"`
}

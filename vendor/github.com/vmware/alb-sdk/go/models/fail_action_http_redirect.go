package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FailActionHTTPRedirect fail action HTTP redirect
// swagger:model FailActionHTTPRedirect
type FailActionHTTPRedirect struct {

	// host of FailActionHTTPRedirect.
	// Required: true
	Host *string `json:"host"`

	// path of FailActionHTTPRedirect.
	Path *string `json:"path,omitempty"`

	//  Enum options - HTTP, HTTPS.
	Protocol *string `json:"protocol,omitempty"`

	// query of FailActionHTTPRedirect.
	Query *string `json:"query,omitempty"`

	//  Enum options - HTTP_REDIRECT_STATUS_CODE_301, HTTP_REDIRECT_STATUS_CODE_302, HTTP_REDIRECT_STATUS_CODE_307.
	StatusCode *string `json:"status_code,omitempty"`
}

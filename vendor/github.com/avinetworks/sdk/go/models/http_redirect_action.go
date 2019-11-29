package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPRedirectAction HTTP redirect action
// swagger:model HTTPRedirectAction
type HTTPRedirectAction struct {

	// Host config.
	Host *URIParam `json:"host,omitempty"`

	// Keep or drop the query of the incoming request URI in the redirected URI.
	KeepQuery *bool `json:"keep_query,omitempty"`

	// Path config.
	Path *URIParam `json:"path,omitempty"`

	// Port to which redirect the request. Allowed values are 1-65535.
	Port *int32 `json:"port,omitempty"`

	// Protocol type. Enum options - HTTP, HTTPS.
	// Required: true
	Protocol *string `json:"protocol"`

	// HTTP redirect status code. Enum options - HTTP_REDIRECT_STATUS_CODE_301, HTTP_REDIRECT_STATUS_CODE_302, HTTP_REDIRECT_STATUS_CODE_307.
	StatusCode *string `json:"status_code,omitempty"`
}

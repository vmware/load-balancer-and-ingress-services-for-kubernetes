package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPRewriteLocHdrAction HTTP rewrite loc hdr action
// swagger:model HTTPRewriteLocHdrAction
type HTTPRewriteLocHdrAction struct {

	// Host config.
	Host *URIParam `json:"host,omitempty"`

	// Keep or drop the query from the server side redirect URI.
	KeepQuery *bool `json:"keep_query,omitempty"`

	// Path config.
	Path *URIParam `json:"path,omitempty"`

	// Port to use in the redirected URI. Allowed values are 1-65535.
	Port *int32 `json:"port,omitempty"`

	// HTTP protocol type. Enum options - HTTP, HTTPS.
	// Required: true
	Protocol *string `json:"protocol"`
}

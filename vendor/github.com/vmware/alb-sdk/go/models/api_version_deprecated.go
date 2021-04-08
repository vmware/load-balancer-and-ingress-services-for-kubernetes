package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// APIVersionDeprecated Api version deprecated
// swagger:model ApiVersionDeprecated
type APIVersionDeprecated struct {

	// API version used. Field introduced in 21.1.1.
	APIVersionUsed *string `json:"api_version_used,omitempty"`

	// IP address of client who sent the request. Field introduced in 21.1.1.
	ClientIP *string `json:"client_ip,omitempty"`

	// Minimum supported API version. Field introduced in 21.1.1.
	MinSupportedAPIVersion *string `json:"min_supported_api_version,omitempty"`

	// URI of the request. Field introduced in 21.1.1.
	Path *string `json:"path,omitempty"`

	// User who sent the request. Field introduced in 21.1.1.
	User *string `json:"user,omitempty"`
}

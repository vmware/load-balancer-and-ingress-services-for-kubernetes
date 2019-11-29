package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MethodMatch method match
// swagger:model MethodMatch
type MethodMatch struct {

	// Criterion to use for HTTP method matching the method in the HTTP request. Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Configure HTTP method(s). Enum options - HTTP_METHOD_GET, HTTP_METHOD_HEAD, HTTP_METHOD_PUT, HTTP_METHOD_DELETE, HTTP_METHOD_POST, HTTP_METHOD_OPTIONS, HTTP_METHOD_TRACE, HTTP_METHOD_CONNECT.
	Methods []string `json:"methods,omitempty"`
}

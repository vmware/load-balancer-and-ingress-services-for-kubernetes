package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPSMMatchElement waf p s m match element
// swagger:model WafPSMMatchElement
type WafPSMMatchElement struct {

	// Mark this element excluded, like in '!ARGS password'. Field introduced in 18.2.3.
	Excluded *bool `json:"excluded,omitempty"`

	// Match_element index. Field introduced in 18.2.3.
	// Required: true
	Index *int32 `json:"index"`

	// The variable specification. For example ARGS or REQUEST_COOKIES. This can be a scalar like PATH_INFO. Enum options - WAF_VARIABLE_ARGS, WAF_VARIABLE_ARGS_GET, WAF_VARIABLE_ARGS_POST, WAF_VARIABLE_ARGS_NAMES, WAF_VARIABLE_REQUEST_COOKIES, WAF_VARIABLE_QUERY_STRING, WAF_VARIABLE_REQUEST_BASENAME, WAF_VARIABLE_REQUEST_URI, WAF_VARIABLE_PATH_INFO. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`

	// The name of the request collection element. This can be empty, if we address the whole collection or a scalar element. Field introduced in 18.2.3.
	SubElement *string `json:"sub_element,omitempty"`
}

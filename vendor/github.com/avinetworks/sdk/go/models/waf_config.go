package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafConfig waf config
// swagger:model WafConfig
type WafConfig struct {

	// WAF allowed HTTP Versions. Enum options - ZERO_NINE, ONE_ZERO, ONE_ONE. Field introduced in 17.2.1.
	AllowedHTTPVersions []string `json:"allowed_http_versions,omitempty"`

	// WAF allowed HTTP methods. Enum options - HTTP_METHOD_GET, HTTP_METHOD_HEAD, HTTP_METHOD_PUT, HTTP_METHOD_DELETE, HTTP_METHOD_POST, HTTP_METHOD_OPTIONS, HTTP_METHOD_TRACE, HTTP_METHOD_CONNECT. Field introduced in 17.2.1.
	AllowedMethods []string `json:"allowed_methods,omitempty"`

	// WAF allowed Content Types. Field introduced in 17.2.1.
	AllowedRequestContentTypes []string `json:"allowed_request_content_types,omitempty"`

	// Argument seperator. Field introduced in 17.2.1.
	ArgumentSeparator *string `json:"argument_separator,omitempty"`

	// Enable to buffer response body for inspection. Field introduced in 17.2.3.
	BufferResponseBodyForInspection *bool `json:"buffer_response_body_for_inspection,omitempty"`

	// Maximum size for the client request body for file uploads. Allowed values are 1-32768. Field deprecated in 18.1.5. Field introduced in 17.2.1.
	ClientFileUploadMaxBodySize *int32 `json:"client_file_upload_max_body_size,omitempty"`

	// Maximum size for the client request body for non-file uploads. Allowed values are 1-32768. Field deprecated in 18.1.5. Field introduced in 17.2.1.
	ClientNonfileUploadMaxBodySize *int32 `json:"client_nonfile_upload_max_body_size,omitempty"`

	// Maximum size for the client request body scanned by WAF. Allowed values are 1-32768. Field introduced in 18.1.5, 18.2.1.
	ClientRequestMaxBodySize *int32 `json:"client_request_max_body_size,omitempty"`

	// 0  For Netscape Cookies. 1  For version 1 cookies. Allowed values are 0-1. Field introduced in 17.2.1.
	CookieFormatVersion *int32 `json:"cookie_format_version,omitempty"`

	// Ignore request body parsing errors due to partial scanning. Field introduced in 18.1.5, 18.2.1.
	IgnoreIncompleteRequestBodyError *bool `json:"ignore_incomplete_request_body_error,omitempty"`

	// The maximum period of time WAF processing is allowed to take for a single request. A value of 0 (zero) means no limit and should not be chosen in production deployments. It is only used for exceptional situations where crashes of se_dp processes are acceptable. The behavior of the system if this time is exceeded depends on two other configuration settings, the WAF policy mode and the WAF failure mode. In WAF policy mode 'Detection', the request is allowed and flagged for both failure mode 'Closed' and 'Open'. In enforcement node, 'Closed' means the request is rejected, 'Open' means the request is allowed and flagged. Irrespective of these settings, no subsequent WAF rules of this or other phases will be executed once the maximum execution time has been exceeded. Allowed values are 0-5000. Field introduced in 17.2.12, 18.1.2.
	MaxExecutionTime *int32 `json:"max_execution_time,omitempty"`

	// Limit CPU utilization for each regular expression match when processing rules. Field introduced in 17.2.5.
	RegexMatchLimit *int32 `json:"regex_match_limit,omitempty"`

	// WAF default action for Request Body Phase. Field introduced in 17.2.1.
	// Required: true
	RequestBodyDefaultAction *string `json:"request_body_default_action"`

	// WAF default action for Request Header Phase. Field introduced in 17.2.1.
	// Required: true
	RequestHdrDefaultAction *string `json:"request_hdr_default_action"`

	// WAF default action for Response Body Phase. Field introduced in 17.2.1.
	// Required: true
	ResponseBodyDefaultAction *string `json:"response_body_default_action"`

	// WAF default action for Response Header Phase. Field introduced in 17.2.1.
	// Required: true
	ResponseHdrDefaultAction *string `json:"response_hdr_default_action"`

	// WAF Restricted File Extensions. Field introduced in 17.2.1.
	RestrictedExtensions []string `json:"restricted_extensions,omitempty"`

	// WAF Restricted HTTP Headers. Field introduced in 17.2.1.
	RestrictedHeaders []string `json:"restricted_headers,omitempty"`

	// Maximum size for response body scanned by WAF. Allowed values are 1-32768. Field introduced in 17.2.1.
	ServerResponseMaxBodySize *int32 `json:"server_response_max_body_size,omitempty"`

	// WAF Static File Extensions. GET and HEAD requests with no query args and one of these extensions are whitelisted and not checked by the ruleset. Field introduced in 17.2.5.
	StaticExtensions []string `json:"static_extensions,omitempty"`
}

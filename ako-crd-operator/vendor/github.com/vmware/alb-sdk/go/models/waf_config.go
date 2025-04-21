// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafConfig waf config
// swagger:model WafConfig
type WafConfig struct {

	// WAF allowed HTTP Versions. Enum options - ZERO_NINE, ONE_ZERO, ONE_ONE, TWO_ZERO. Field introduced in 17.2.1. Maximum of 8 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowedHTTPVersions []string `json:"allowed_http_versions,omitempty"`

	// WAF allowed HTTP methods. Enum options - HTTP_METHOD_GET, HTTP_METHOD_HEAD, HTTP_METHOD_PUT, HTTP_METHOD_DELETE, HTTP_METHOD_POST, HTTP_METHOD_OPTIONS, HTTP_METHOD_TRACE, HTTP_METHOD_CONNECT, HTTP_METHOD_PATCH, HTTP_METHOD_PROPFIND, HTTP_METHOD_PROPPATCH, HTTP_METHOD_MKCOL, HTTP_METHOD_COPY, HTTP_METHOD_MOVE, HTTP_METHOD_LOCK, HTTP_METHOD_UNLOCK. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowedMethods []string `json:"allowed_methods,omitempty"`

	// Allowed request content type character sets in WAF. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AllowedRequestContentTypeCharsets []string `json:"allowed_request_content_type_charsets,omitempty"`

	// Argument seperator. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ArgumentSeparator *string `json:"argument_separator,omitempty"`

	// Maximum size for the client request body scanned by WAF. Allowed values are 1-32768. Field introduced in 18.1.5, 18.2.1. Unit is KB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientRequestMaxBodySize *uint32 `json:"client_request_max_body_size,omitempty"`

	// WAF Content-Types and their request body parsers. Use this field to configure which Content-Types should be handled by WAF and which parser should be used. All Content-Types here are treated as 'allowed'. The order of entries matters. If the request's Content-Type matches an entry, its request body parser will run and no other parser will be invoked. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ContentTypeMappings []*WafContentTypeMapping `json:"content_type_mappings,omitempty"`

	// 0  For Netscape Cookies. 1  For version 1 cookies. Allowed values are 0-1. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CookieFormatVersion uint32 `json:"cookie_format_version,omitempty"`

	// Ignore request body parsing errors due to partial scanning. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IgnoreIncompleteRequestBodyError *bool `json:"ignore_incomplete_request_body_error,omitempty"`

	// The maximum period of time WAF processing is allowed to take for a single request. A value of 0 (zero) means no limit and should not be chosen in production deployments. It is only used for exceptional situations where crashes of se_dp processes are acceptable. The behavior of the system if this time is exceeded depends on two other configuration settings, the WAF policy mode and the WAF failure mode. In WAF policy mode 'Detection', the request is allowed and flagged for both failure mode 'Closed' and 'Open'. In enforcement node, 'Closed' means the request is rejected, 'Open' means the request is allowed and flagged. Irrespective of these settings, no subsequent WAF rules of this or other phases will be executed once the maximum execution time has been exceeded. Allowed values are 0-5000. Field introduced in 17.2.12, 18.1.2. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxExecutionTime *uint32 `json:"max_execution_time,omitempty"`

	// Limit CPU utilization for each regular expression match when processing rules. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RegexMatchLimit *uint32 `json:"regex_match_limit,omitempty"`

	// Limit depth of recursion for each regular expression match when processing rules. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RegexRecursionLimit *uint32 `json:"regex_recursion_limit,omitempty"`

	// WAF default action for Request Body Phase. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RequestBodyDefaultAction *string `json:"request_body_default_action"`

	// WAF default action for Request Header Phase. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RequestHdrDefaultAction *string `json:"request_hdr_default_action"`

	// WAF default action for Response Body Phase. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ResponseBodyDefaultAction *string `json:"response_body_default_action"`

	// WAF default action for Response Header Phase. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ResponseHdrDefaultAction *string `json:"response_hdr_default_action"`

	// WAF Restricted File Extensions. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RestrictedExtensions []string `json:"restricted_extensions,omitempty"`

	// WAF Restricted HTTP Headers. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RestrictedHeaders []string `json:"restricted_headers,omitempty"`

	// Whether or not to send WAF status in a request header to pool servers. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SendStatusHeader *bool `json:"send_status_header,omitempty"`

	// Maximum size for response body scanned by WAF. Allowed values are 1-32768. Field introduced in 17.2.1. Unit is KB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerResponseMaxBodySize *uint32 `json:"server_response_max_body_size,omitempty"`

	// WAF Static File Extensions. GET and HEAD requests with no query args and one of these extensions are allowed and not checked by the ruleset. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StaticExtensions []string `json:"static_extensions,omitempty"`

	// HTTP status code used by WAF Positive Security Model when rejecting a request. Enum options - HTTP_RESPONSE_CODE_0, HTTP_RESPONSE_CODE_100, HTTP_RESPONSE_CODE_101, HTTP_RESPONSE_CODE_200, HTTP_RESPONSE_CODE_201, HTTP_RESPONSE_CODE_202, HTTP_RESPONSE_CODE_203, HTTP_RESPONSE_CODE_204, HTTP_RESPONSE_CODE_205, HTTP_RESPONSE_CODE_206, HTTP_RESPONSE_CODE_300, HTTP_RESPONSE_CODE_301, HTTP_RESPONSE_CODE_302, HTTP_RESPONSE_CODE_303, HTTP_RESPONSE_CODE_304, HTTP_RESPONSE_CODE_305, HTTP_RESPONSE_CODE_307, HTTP_RESPONSE_CODE_400, HTTP_RESPONSE_CODE_401, HTTP_RESPONSE_CODE_402.... Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StatusCodeForRejectedRequests *string `json:"status_code_for_rejected_requests,omitempty"`

	// The name of the request header indicating WAF evaluation status to pool servers. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StatusHeaderName *string `json:"status_header_name,omitempty"`

	// Block or flag XML requests referring to External Entities. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	XMLXxeProtection *bool `json:"xml_xxe_protection,omitempty"`
}

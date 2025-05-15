/*
Copyright 2019-2025 VMware, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationProfileType defines the type of application profile.
// +kubebuilder:validation:Enum=ApplicationProfileHTTP
type ApplicationProfileType string

const (
	// types of application profile

	// HTTP application proxy is enabled for this virtual service.
	ApplicationProfileHTTP ApplicationProfileType = "HTTP"
)

// HTTPCookieSameSite defines the SameSite attribute for HTTP cookies.
// +kubebuilder:validation:Enum=SAMESITE_NONE;SAMESITE_LAX;SAMESITE_STRICT
type HTTPCookieSameSite string

const (
	//HTTP cookie SameSite attribute value None.
	SAMESITE_NONE HTTPCookieSameSite = "None"
	//HTTP cookie SameSite attribute value Lax.
	SAMESITE_LAX HTTPCookieSameSite = "Lax"
	//HTTP cookie SameSite attribute value Strict.
	SAMESITE_STRICT HTTPCookieSameSite = "Strict"
)

// SSLClientCertificateMode defines the client-side SSL verification mode.
// +kubebuilder:validation:Enum=SSL_CLIENT_CERTIFICATE_NONE;SSL_CLIENT_CERTIFICATE_REQUEST;SSL_CLIENT_CERTIFICATE_REQUIRE
type SSLClientCertificateMode string

const (
	SSL_CLIENT_CERTIFICATE_NONE    SSLClientCertificateMode = "SSL_CLIENT_CERTIFICATE_NONE"
	SSL_CLIENT_CERTIFICATE_REQUEST SSLClientCertificateMode = "SSL_CLIENT_CERTIFICATE_REQUEST"
	SSL_CLIENT_CERTIFICATE_REQUIRE SSLClientCertificateMode = "SSL_CLIENT_CERTIFICATE_REQUIRE"
)

// AppServiceType defines the application service type.
// +kubebuilder:validation:Enum=APP_SERVICE_TYPE_L7_HORIZON;APP_SERVICE_TYPE_L4_BLAST;APP_SERVICE_TYPE_L4_PCOIP;APP_SERVICE_TYPE_L4_FTP
type AppServiceType string

const (
	// Layer 7 application proxy and Horizon app.
	APP_SERVICE_TYPE_L7_HORIZON AppServiceType = "HORIZON"
	// Layer 4 application proxy and protocol blast.
	APP_SERVICE_TYPE_L4_BLAST AppServiceType = "BLAST"
	// Layer 4 application proxy and protocol PCoIP.
	APP_SERVICE_TYPE_L4_PCOIP AppServiceType = "PCoIP"
	// Layer 4 application proxy and protocol FTP.
	APP_SERVICE_TYPE_L4_FTP AppServiceType = "FTP"
)

// ProxyProtocolVersion defines the Proxy Protocol version.
// +kubebuilder:validation:Enum=PROXY_PROTOCOL_VERSION_1;PROXY_PROTOCOL_VERSION_2
type ProxyProtocolVersion string

const (
	//Proxy Protocol Version 1
	PROXY_PROTOCOL_VERSION_1 ProxyProtocolVersion = "ProxyProtocolV1"
	//Proxy Protocol Version 2
	PROXY_PROTOCOL_VERSION_2 ProxyProtocolVersion = "ProxyProtocolV2"
)

// XFFUpdate defines how X-Forwarded-For headers are handled.
// +kubebuilder:validation:Enum=REPLACE_XFF_HEADERS;APPEND_TO_THE_XFF_HEADER;ADD_NEW_XFF_HEADER
type XFFUpdate string

const (
	//Drop all the incoming xff headers and add a new one with client's IP address
	REPLACE_XFF_HEADERS XFFUpdate = "Replace all incoming X-Forward-For headers with the Avi created header."
	//Appends all the incoming XFF headers and client's IP address together
	APPEND_TO_THE_XFF_HEADER XFFUpdate = "All incoming X-Forwarded-For headers will be appended to the Avi created header."
	//Add new XFF header with client's IP address
	ADD_NEW_XFF_HEADER XFFUpdate = "Simply add a new X-Forwarded-For header."
)

// TrueClientIPIndexDirection defines the direction to count IPs in the header.
// +kubebuilder:validation:Enum=LEFT;RIGHT
type TrueClientIPIndexDirection string

const (
	//From Left.
	LEFT TrueClientIPIndexDirection = "From Left."
	//From Right.
	RIGHT TrueClientIPIndexDirection = "From Right."
)

// ReqRateLimitType defines the type of request rate limiting.
// +kubebuilder:validation:Enum=RL_CLIENT_IP;RL_URI;RL_CLIENT_IP_URI;RL_CLIENT_IP_BAD;RL_URI_BAD;RL_CLIENT_IP_URI_BAD;RL_CLIENT_IP_SCAN;RL_URI_SCAN;RL_CONN;RL_REQ;RL_HEADER;RL_CUSTOM
type ReqRateLimitType string

const (
	RL_CLIENT_IP         ReqRateLimitType = "RL_CLIENT_IP"         // Per client IP.
	RL_URI               ReqRateLimitType = "RL_URI"               // Per URI.
	RL_CLIENT_IP_URI     ReqRateLimitType = "RL_CLIENT_IP_URI"     // Per client IP and URI.
	RL_CLIENT_IP_BAD     ReqRateLimitType = "RL_CLIENT_IP_BAD"     // Per client IP.
	RL_URI_BAD           ReqRateLimitType = "RL_URI_BAD"           // Per URI.
	RL_CLIENT_IP_URI_BAD ReqRateLimitType = "RL_CLIENT_IP_URI_BAD" // Per client IP and URI.
	RL_CLIENT_IP_SCAN    ReqRateLimitType = "RL_CLIENT_IP_SCAN"    // Client IP scanning.
	RL_URI_SCAN          ReqRateLimitType = "RL_URI_SCAN"          // URI scanning.
	RL_CONN              ReqRateLimitType = "RL_CONN"              // Connection.
	RL_REQ               ReqRateLimitType = "RL_REQ"               // Request.
	RL_HEADER            ReqRateLimitType = "RL_HEADER"            // HTTP header.
	RL_CUSTOM            ReqRateLimitType = "RL_CUSTOM"            // Custom string.
)

// RateLimiterActionType defines the type of action to take when rate limiting.
// +kubebuilder:validation:Enum=RL_ACTION_NONE;RL_ACTION_DROP_CONN;RL_ACTION_RESET_CONN;RL_ACTION_CLOSE_CONN;RL_ACTION_LOCAL_RSP;RL_ACTION_REDIRECT
type RateLimiterActionType string

const (
	//No action for rate limiting.
	RL_ACTION_NONE RateLimiterActionType = "RL_ACTION_NONE"
	//Drop rate limited syns.
	RL_ACTION_DROP_CONN RateLimiterActionType = "RL_ACTION_DROP_CONN"
	//Drop rate limited syns and send a reset in case of TCP.
	RL_ACTION_RESET_CONN RateLimiterActionType = "RL_ACTION_RESET_CONN"
	//Close connection when requests get rate limited.
	RL_ACTION_CLOSE_CONN RateLimiterActionType = "RL_ACTION_CLOSE_CONN"
	//Send local response for rate limited requests.
	RL_ACTION_LOCAL_RSP RateLimiterActionType = "RL_ACTION_LOCAL_RSP"
	//Redirect rate limited requests.
	RL_ACTION_REDIRECT RateLimiterActionType = "RL_ACTION_REDIRECT"
)

// +kubebuilder:validation:Enum=SENSITIVE;INSENSITIVE
type MatchCase string

const (
	// case sensitive match
	SENSITIVE MatchCase = "Sensitive"
	// case insensitive match
	INSENSITIVE MatchCase = "Insensitive"
)

// +kubebuilder:validation:Enum=BEGINS_WITH;DOES_NOT_BEGIN_WITH;CONTAINS;DOES_NOT_CONTAIN;ENDS_WITH;DOES_NOT_END_WITH;EQUALS;DOES_NOT_EQUAL;REGEX_MATCH;REGEX_DOES_NOT_MATCH
type StringOperation string

const (
	// begins with the configured value(s)
	BEGINS_WITH StringOperation = "Begins with"
	// does not begin with the configured value(s)
	DOES_NOT_BEGIN_WITH StringOperation = "Does not begin with"
	// contains the configured value(s)
	CONTAINS StringOperation = "Contains"
	// does not contain the configured value(s)
	DOES_NOT_CONTAIN StringOperation = "Does not contain"
	// ends with the configured value(s)
	ENDS_WITH StringOperation = "Ends with"
	// does not end with the configured value(s)
	DOES_NOT_END_WITH StringOperation = "Does not end with"
	// equals the configured value(s)
	EQUALS StringOperation = "Equals"
	// does not equal the configured value(s)
	DOES_NOT_EQUAL StringOperation = "Does not equal"
	// regex pattern matches with the configured value(s)
	REGEX_MATCH StringOperation = "Regex pattern matches"
	// regex pattern does not match with the configured value(s)
	REGEX_DOES_NOT_MATCH StringOperation = "Regex pattern does not match"
)

// HTTPProtocol defines the protocol for HTTP redirection.
// +kubebuilder:validation:Enum=HTTP;HTTPS
type HTTPProtocol string

const (
	// HTTP Protocol
	HTTP HTTPProtocol = "HTTP"
	// Secure HTTP Protocol
	HTTPS HTTPProtocol = "HTTPS"
)

// Host, Path and Query can be specified either as a full string or as a
// composition of tokens from the three pieces.
// Examples:
// Host can be set to www.avinetworks.com
// -OR-
// If the URI is www.avinetworks.com/sales/blah, the rewrite can make it
// sales.avinetworks.com/blah, which will specify that the rewrite parameters
// are :
// host = path[0].host[1:] and path = path[1:]
// +kubebuilder:validation:Enum=URI_PARAM_TYPE_TOKENIZED
type URIParamType string

const (
	URI_PARAM_TYPE_TOKENIZED URIParamType = "URI_PARAM_TYPE_TOKENIZED"
)

// +kubebuilder:validation:Enum=URI_TOKEN_TYPE_HOST;URI_TOKEN_TYPE_PATH;URI_TOKEN_TYPE_STRING;URI_TOKEN_TYPE_STRING_GROUP;URI_TOKEN_TYPE_REGEX;URI_TOKEN_TYPE_REGEX_QUERY
type URITokenType string

const (
	// Use host component of the URI as tokens
	URI_TOKEN_TYPE_HOST URITokenType = "URI_TOKEN_TYPE_HOST"
	// Use path component of the URI as tokens
	URI_TOKEN_TYPE_PATH URITokenType = "URI_TOKEN_TYPE_PATH"
	// Use constant string as a token
	URI_TOKEN_TYPE_STRING URITokenType = "URI_TOKEN_TYPE_STRING"
	// Use result of StringGroup lookup as a token
	URI_TOKEN_TYPE_STRING_GROUP URITokenType = "URI_TOKEN_TYPE_STRING_GROUP"
	// Use regex captures from URI path as a token.
	URI_TOKEN_TYPE_REGEX URITokenType = "URI_TOKEN_TYPE_REGEX"
	// Use regex captures from URI query as a token.
	URI_TOKEN_TYPE_REGEX_QUERY URITokenType = "URI_TOKEN_TYPE_REGEX_QUERY"
)

// HTTPRedirectStatusCode defines the HTTP redirect status code.
// +kubebuilder:validation:Enum=HTTP_REDIRECT_STATUS_CODE_301;HTTP_REDIRECT_STATUS_CODE_302;HTTP_REDIRECT_STATUS_CODE_307
type HTTPRedirectStatusCode string

const (
	//301 Moved Permanently
	HTTP_REDIRECT_STATUS_CODE_301 HTTPRedirectStatusCode = "301"
	//302 Found
	HTTP_REDIRECT_STATUS_CODE_302 HTTPRedirectStatusCode = "302"
	//307 Moved Temporarily (only for HTTP/1.1)
	HTTP_REDIRECT_STATUS_CODE_307 HTTPRedirectStatusCode = "307"
)

// HTTPLocalResponseStatusCode defines HTTP status code for local response rate limit action.
// +kubebuilder:validation:Enum=HTTP_LOCAL_RESPONSE_STATUS_CODE_200;HTTP_LOCAL_RESPONSE_STATUS_CODE_204;HTTP_LOCAL_RESPONSE_STATUS_CODE_403;HTTP_LOCAL_RESPONSE_STATUS_CODE_404;HTTP_LOCAL_RESPONSE_STATUS_CODE_429;HTTP_LOCAL_RESPONSE_STATUS_CODE_501
type HTTPLocalResponseStatusCode string

const (
	HTTP_LOCAL_RESPONSE_STATUS_CODE_200 HTTPLocalResponseStatusCode = "200" //200 OK
	HTTP_LOCAL_RESPONSE_STATUS_CODE_204 HTTPLocalResponseStatusCode = "204" //204 No Content
	HTTP_LOCAL_RESPONSE_STATUS_CODE_403 HTTPLocalResponseStatusCode = "403" //403 Forbidden
	HTTP_LOCAL_RESPONSE_STATUS_CODE_404 HTTPLocalResponseStatusCode = "404" //404 Not Found
	HTTP_LOCAL_RESPONSE_STATUS_CODE_429 HTTPLocalResponseStatusCode = "429" //429 Too Many Requests (RFC 6585)
	HTTP_LOCAL_RESPONSE_STATUS_CODE_501 HTTPLocalResponseStatusCode = "501" //501 Not Implemented
)

// HTTPPolicyVar defines a variable used in HTTP policies.
// +kubebuilder:validation:Enum=HTTP_POLICY_VAR_CLIENT_IP;HTTP_POLICY_VAR_VS_PORT;HTTP_POLICY_VAR_VS_IP;HTTP_POLICY_VAR_HTTP_HDR;HTTP_POLICY_VAR_SSL_CLIENT_FINGERPRINT;HTTP_POLICY_VAR_SSL_CLIENT_SERIAL;HTTP_POLICY_VAR_SSL_CLIENT_ISSUER;HTTP_POLICY_VAR_SSL_CLIENT_SUBJECT;HTTP_POLICY_VAR_SSL_CLIENT_RAW;HTTP_POLICY_VAR_SSL_PROTOCOL;HTTP_POLICY_VAR_SSL_SERVER_NAME;HTTP_POLICY_VAR_USER_NAME;HTTP_POLICY_VAR_SSL_CIPHER;HTTP_POLICY_VAR_REQUEST_ID;HTTP_POLICY_VAR_SSL_CLIENT_VERSION;HTTP_POLICY_VAR_SSL_CLIENT_SIGALG;HTTP_POLICY_VAR_SSL_CLIENT_NOTVALIDBEFORE;HTTP_POLICY_VAR_SSL_CLIENT_NOTVALIDAFTER;HTTP_POLICY_VAR_SSL_CLIENT_ESCAPED;HTTP_POLICY_VAR_SOURCE_IP
type HTTPPolicyVar string

const (
	// Variable to get client IP address of the request
	HTTP_POLICY_VAR_CLIENT_IP HTTPPolicyVar = "HTTP_POLICY_VAR_CLIENT_IP"
	// Variable to get virtual service port the request came on
	HTTP_POLICY_VAR_VS_PORT HTTPPolicyVar = "HTTP_POLICY_VAR_VS_PORT"
	// Variable to get IP address of the virtual service the request came for
	HTTP_POLICY_VAR_VS_IP HTTPPolicyVar = "HTTP_POLICY_VAR_VS_IP"
	// Variable to get value of the HTTP header in the HTTP request
	HTTP_POLICY_VAR_HTTP_HDR HTTPPolicyVar = "HTTP_POLICY_VAR_HTTP_HDR"
	// SSL Client Certificate fingerprint
	HTTP_POLICY_VAR_SSL_CLIENT_FINGERPRINT HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_CLIENT_FINGERPRINT"
	// SSL Client Certificate serial number
	HTTP_POLICY_VAR_SSL_CLIENT_SERIAL HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_CLIENT_SERIAL"
	// SSL Client Certificate Issuer name
	HTTP_POLICY_VAR_SSL_CLIENT_ISSUER HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_CLIENT_ISSUER"
	// SSL Client Certificate Subject name
	HTTP_POLICY_VAR_SSL_CLIENT_SUBJECT HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_CLIENT_SUBJECT"
	// The whole SSL client certificate in a PEM format
	HTTP_POLICY_VAR_SSL_CLIENT_RAW HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_CLIENT_RAW"
	// SSL protocol negotiated for this connection
	HTTP_POLICY_VAR_SSL_PROTOCOL HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_PROTOCOL"
	// Server name requested through SNI
	HTTP_POLICY_VAR_SSL_SERVER_NAME HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_SERVER_NAME"
	// Username obtained from http request using basic auth
	HTTP_POLICY_VAR_USER_NAME HTTPPolicyVar = "HTTP_POLICY_VAR_USER_NAME"
	// Cipher suite used for the SSL/TLS connection between the client and virtual service
	HTTP_POLICY_VAR_SSL_CIPHER HTTPPolicyVar = "HTTP_POLICY_VAR_SSL_CIPHER"
	// Variable to get request id of the request
	HTTP_POLICY_VAR_REQUEST_ID HTTPPolicyVar = "HTTP_POLICY_VAR_REQUEST_ID"
	// SSL Client Certificate version.
	HTTP_POLICY_VAR_SSL_CLIENT_VERSION HTTPPolicyVar = "SSL Client Version."
	// SSL Client Certificate signature algorithm.
	HTTP_POLICY_VAR_SSL_CLIENT_SIGALG HTTPPolicyVar = "SSL Client Signature Algorithm."
	// SSL Client Certificate not before date.
	HTTP_POLICY_VAR_SSL_CLIENT_NOTVALIDBEFORE HTTPPolicyVar = "SSL Client Not Valid Before."
	// SSL Client Certificate not after date.
	HTTP_POLICY_VAR_SSL_CLIENT_NOTVALIDAFTER HTTPPolicyVar = "SSL Client Not Valid After."
	// The whole SSL Client Certificate in PEM format and percent encoded.
	HTTP_POLICY_VAR_SSL_CLIENT_ESCAPED HTTPPolicyVar = "SSL Client Certificate Escaped."
	// Variable to get source IP address of the request.
	HTTP_POLICY_VAR_SOURCE_IP HTTPPolicyVar = "Source IP of the client connection."
)

// +kubebuilder:validation:Enum=IS_IN;IS_NOT_IN
type MatchOperation string

const (
	// is in the configured value(s)
	IS_IN MatchOperation = "Is in."
	// is not in the configured value(s)
	IS_NOT_IN MatchOperation = "Is not in"
)

type SSLClientRequestHeader struct {

	// If this header exists, reset the connection. If the ssl variable is specified, add a header with this value"
	// +optional
	RequestHeader string `json:"request_header,omitempty"`
	// "Set the request header with the value as indicated by this SSL variable. Eg. send the whole certificate in PEM format"
	// +optional
	RequestHeaderValue HTTPPolicyVar `json:"request_header_value,omitempty"`
}

// We are adding the SSL client certificate specific actions in
// the parameters below. At this time, we only support a restricted
// set of match criteria which is to match if the header exists and the
// only action is to reset the connection. Based on how this evolves,
// this protobuf will be enhanced. Internally, this will be converted to an
// internal policy
type SSLClientCertificateAction struct {
	Headers []SSLClientRequestHeader `json:"headers,omitempty"`
	// +optional
	// +kubebuilder:default:=false
	CloseConnection bool `json:"close_connection,omitempty"`
}

type PathMatch struct {
	// Criterion to use for matching the path in the HTTP request URI.
	// +kubebuilder:default:="CONTAINS"
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=BEGINS_WITH;DOES_NOT_BEGIN_WITH;CONTAINS;DOES_NOT_CONTAIN;ENDS_WITH;DOES_NOT_END_WITH;EQUALS;DOES_NOT_EQUAL
	MatchCriteria StringOperation `json:"match_criteria"`

	// Case sensitivity to use for the matching
	// +kubebuilder:default:="INSENSITIVE"
	// +optional
	// Match case
	MatchCase MatchCase `json:"match_case"`

	// String values
	// Implicit Path Group
	// TODO: (f_skip_input_check) =  true
	MatchStr []string `json:"match_str,omitempty"`

	// UUID of the string group(s)
	// User-created named Path Group
	// +kubebuilder:validation:Reference:type=StringGroup
	StringGroupUuids []string `json:"string_group_uuids,omitempty"`

	// Match against the decoded URI path.
	// +kubebuilder:default:=true
	// +optional
	MatchDecodedString bool `json:"match_decoded_string,omitempty"`
}

// HttpCacheConfig defines HTTP caching configuration.
type HttpCacheConfig struct {
	// +kubebuilder:default:=false
	// +optional
	// Enable/disable HTTP object caching. When enabling caching for the first time, SE Group app_cache_percent must be set to allocate shared memory required for caching (A service engine restart is needed after setting/resetting the SE group value).
	Enabled bool `json:"enabled"`
	// +kubebuilder:default:=true
	// +optional
	// Add an X-Cache header to content served from cache, which indicates to the client that the object was served from an intermediate cache.
	XcacheHeader bool `json:"xcache_header"`
	// +kubebuilder:default:=true
	// +optional
	// Add an Age header to content served from cache, which indicates to the client the number of seconds the object has been in the cache.
	AgeHeader bool `json:"age_header"`
	// +kubebuilder:default:=true
	// +optional
	// If a Date header was not added by the server, add a Date header to the object served from cache. This indicates to the client when the object was originally sent by the server to the cache.
	DateHeader bool `json:"date_header"`
	// +kubebuilder:default:=100
	// +optional
	// Minimum size of an object to store in the cache.
	MinObjectSize uint32 `json:"min_object_size"`
	// +kubebuilder:default:=4194304
	// +optional
	// Maximum size of an object to store in the cache.
	MaxObjectSize uint32 `json:"max_object_size"`
	// +kubebuilder:default:=600
	// +optional
	// Default expiration time of cache objects received from the server without a Cache-Control expiration header. This value may be overwritten by the Heuristic Expire setting.
	DefaultExpire uint32 `json:"default_expire"`
	// +kubebuilder:default:=false
	// +optional
	// If a response object from the server does not include the Cache-Control header, but does include a Last-Modified header, the system will use this time to calculate the Cache-Control expiration. If unable to solicit an Last-Modified header, then the system will fall back to the Cache Expire Time value.
	HeuristicExpire bool `json:"heuristic_expire"`
	// +kubebuilder:default:=0
	// +optional
	// Max size, in bytes, of the cache. The default, zero, indicates auto configuration.
	MaxCacheSize uint64 `json:"max_cache_size"`
	// +kubebuilder:default:=false
	// +optional
	// Allow caching of objects whose URI included a query argument. When disabled, these objects are not cached. When enabled, the request must match the URI query to be considered a hit.
	QueryCacheable bool `json:"query_cacheable"`
	// Allowlist of cacheable mime types. If both Cacheable Mime Types string list and string group are empty, this defaults to */*
	// TODO (f_auto_enforce) = "STR_CACHE_MIME"
	MimeTypesList []string `json:"mime_types_list,omitempty"`
	// Allowlist string group of cacheable mime types. If both Cacheable Mime Types string list and string group are empty, this defaults to */*
	// TODO (f_auto_enforce) = "STRGROUPS_CACHE_MIME"
	// +kubebuilder:validation:Reference:type=StringGroup
	MimeTypesGroupUuids []string `json:"mime_types_group_uuids,omitempty"`
	// +kubebuilder:default:=false
	// Enable/disable caching objects without Cache-Control headers
	// +optional
	Aggressive bool `json:"aggressive"`
	// Non-cacheable URI configuration with match criteria.
	// +optional
	UriNonCacheable *PathMatch `json:"uri_non_cacheable,omitempty"`
	// +kubebuilder:default:=false
	// Ignore client's cache control headers when fetching or storing from and to the cache.
	// +optional
	IgnoreRequestCacheControl bool `json:"ignore_request_cache_control"`
	// Blocklist of non-cacheable mime types
	// TODO (f_auto_enforce) = "STR_NO_CACHE_MIME"
	MimeTypesBlockLists []string `json:"mime_types_block_lists,omitempty"`
	// Blocklist string group of non-cacheable mime types
	// TODO (f_auto_enforce) = "STRGROUPS_NO_CACHE_MIME"
	// +kubebuilder:validation:Reference:type=StringGroup
	MimeTypesBlockGroupUuids []string `json:"mime_types_block_group_uuids,omitempty"`
}

// RateProfile defines rate limiting parameters.
type RateProfile struct {

	// +optional
	// +kubebuilder:default:=false
	// +kubebuilder:validation:XValidation:rule="self == false",message="explicit_tracking must be false"
	// TODO if BASIC tier allowed value must be false
	// Explicitly tracks an attacker across rate periods
	ExplicitTracking *bool `json:"explicit_tracking,omitempty"`

	// +optional
	// +kubebuilder:default:=false
	// +kubebuilder:validation:XValidation:rule="self == false",message="fine_grain must be false"
	// TODO if BASIC tier allowed value must be false
	// Enable fine granularity
	FineGrain *bool `json:"fine_grain,omitempty"`

	// +kubebuilder:validation:Required
	// Action to perform upon rate limiting
	// TODO (f_mandatory) = true
	Action *RateLimiterAction `json:"action,omitempty"`

	// +optional
	// HTTP header name.
	// TODO if BASIC tier allow_none: true,
	HTTPHeader *string `json:"http_header,omitempty"`

	// +optional
	// HTTP cookie name.
	// TODO if BASIC tier allow_none: true,
	HTTPCookie *string `json:"http_cookie,omitempty"`
	// +optional
	// The rate limiter configuration for this rate profile.
	RateLimiter *RateLimiter `json:"rate_limiter,omitempty"`
}

// RateLimiter defines the rate limiter configuration.
type RateLimiter struct {
	// +kubebuilder:default:=1000000000
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000000000
	// +optional
	// TODO (f_mandatory) = true
	// Maximum number of connections, requests or packets permitted each period."
	Count *uint32 `json:"count,omitempty"`

	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000000000
	// +optional
	// Time value in seconds to enforce rate count.
	// (units) = SEC
	// TODO (f_mandatory) = true
	// TODO in BASIC tier allowed value is 1
	Period *uint32 `json:"period,omitempty"`

	// +kubebuilder:default:=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000000000
	// Maximum number of connections, requests or packets to be let through instantaneously.  If this is less than count, it will have no effect.
	// TODO in BASIC tier allowed value is 0
	// +optional
	BurstSz *uint32 `json:"burst_sz,omitempty"`

	// Identifier for Rate Limit. Constructed according to context.
	// +optional
	Name *string `json:"name,omitempty"`
}

// RateLimiterAction defines the action to take when rate limiting.
type RateLimiterAction struct {
	// +kubebuilder:default:="RL_ACTION_NONE"
	// +kubebuilder:validation:Enum=RL_ACTION_NONE;RL_ACTION_DROP_CONN
	// Type of action to be enforced upon hitting the rate limit.
	// +optional
	// TODO in BASIC tier only RL_ACTION_NONE,RL_ACTION_DROP_CONN values are allowed
	Type *RateLimiterActionType `json:"type,omitempty"`

	// +optional
	// Parameters for HTTP Redirect rate limit action
	// In the BASIC tier allow_none: true,
	Redirect *HTTPRedirectAction `json:"redirect,omitempty"`

	// +kubebuilder:default:="HTTP_LOCAL_RESPONSE_STATUS_CODE_429"
	// HTTP status code for Local Response rate limit action.
	// TODO In BASIC tier allowed_value: "HTTP_LOCAL_RESPONSE_STATUS_CODE_429"
	StatusCode *HTTPLocalResponseStatusCode `json:"status_code,omitempty"`

	// +optional
	// File to be used for HTTP Local response rate limit action
	// TODO IN BASIC tier allow_none: true
	File *HTTPLocalFile `json:"file,omitempty"`
}

// URIParamToken defines a token within a URI parameter.
// TODO option (m_default) = "URIParamTokenDefault";
type URIParamToken struct {
	// +kubebuilder:validation:Required
	// Token type for constructing the URI
	Type URITokenType `json:"type"`

	// +optional
	// Index of the starting token in the incoming URI
	StartIndex uint32 `json:"start_index,omitempty"`

	// Index of the ending token in the incoming URI
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// TODO (special_values) = "{\"65535\" : \"end of string\"}",
	// +optional
	EndIndex uint32 `json:"end_index,omitempty"`

	// Constant string to use as a token
	// +optional
	StrValue string `json:"str_value,omitempty"`
}

// URIParam defines parameters for URI manipulation in HTTP redirection.
type URIParam struct {
	// +kubebuilder:validation:Required
	// URI param type"
	Type URIParamType `json:"type"`

	// Token config either for the URI components or a constant string
	Tokens []URIParamToken `json:"tokens,omitempty"`
}

// HTTPRedirectAction defines parameters for HTTP redirect rate limit action.
type HTTPRedirectAction struct {
	// +kubebuilder:validation:Required
	// Protocol type
	Protocol HTTPProtocol `json:"protocol"`

	// Host config
	// +optional
	Host *URIParam `json:"host,omitempty"`

	// Port to which redirect the request
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	Port *uint32 `json:"port,omitempty"`

	// Path config
	// +optional
	Path *URIParam `json:"path,omitempty"`

	// +kubebuilder:default:=true
	// +optional
	// Keep or drop the query of the incoming request URI in the redirected URI
	KeepQuery bool `json:"keep_query,omitempty"`

	// +kubebuilder:default:="HTTP_REDIRECT_STATUS_CODE_302"
	// HTTP redirect status code
	// +optional
	StatusCode HTTPRedirectStatusCode `json:"status_code,omitempty"`
	// +optional
	//Add a query string to the redirect URI. If keep_query is set, concatenates the add_string to the query of the incoming request.",
	AddString string `json:"add_string,omitempty"`
}

// HTTPLocalFile defines file to be used for HTTP local response rate limit action.
// TODO option (m_default) = "HTTPLocalFileLength";
type HTTPLocalFile struct {
	// +kubebuilder:validation:Required
	// Mime-type of the content in the file.
	ContentType string `json:"content_type"`

	// +kubebuilder:validation:Required
	// File content to used in the local HTTP response body.
	// TODO (f_skip_input_check) = true
	FileContent string `json:"file_content"`
}

// RateLimiterProfile defines rate limiting settings.
type RateLimiterProfile struct {
	// +optional
	//Rate Limit all connections made from any single client IP address to the Virtual Service.
	ClientIPConnectionsRateLimit *RateProfile `json:"client_ip_connections_rate_limit,omitempty"`
	// +optional
	//Rate Limit all HTTP requests from any single client IP address to all URLs of the Virtual Service.
	ClientIPRequestsRateLimit *RateProfile `json:"client_ip_requests_rate_limit,omitempty"`
	// +optional
	//Rate Limit all HTTP requests from all client IP addresses to any single URL.
	URIRequestsRateLimit *RateProfile `json:"uri_requests_rate_limit,omitempty"`
	// +optional
	//Rate Limit all HTTP requests from any single client IP address to any single URL.
	ClientIPToURIRequestsRateLimit *RateProfile `json:"client_ip_to_uri_requests_rate_limit,omitempty"`
	// +optional
	//Rate Limit all requests from a client for a specified period of time once the count of failed requests from that client crosses a threshold for that period. Clients are tracked based on their IP address. Count and time period are specified through the RateProfile. "Requests are deemed failed based on client or server side error status codes, "consistent with how Avi Logs and Metrics subsystems mark failed requests.
	ClientIPFailedRequestsRateLimit *RateProfile `json:"client_ip_failed_requests_rate_limit,omitempty"`
	// Rate Limit all requests to a URI for a specified
	// period of time once the count of failed requests to that URI crosses a threshold
	// for that period. Count and time period are specified through the RateProfile.
	// Requests are deemed failed based on client or server side error status codes,
	// consistent with how Avi Logs and Metrics subsystems mark failed requests.
	// +optional
	URIFailedRequestsRateLimit *RateProfile `json:"uri_failed_requests_rate_limit,omitempty"`
	// Rate Limit all requests from a client to a URI for a specified
	// period of time once the count of failed requests from that client to the URI crosses a threshold
	// for that period. Clients are tracked based on their IP address.
	// Count and time period are specified through the RateProfile.
	// Requests are deemed failed based on client or server side error status codes,
	// consistent with how Avi Logs and Metrics subsystems mark failed requests.
	// +optional
	ClientIPToURIFailedRequestsRateLimit *RateProfile `json:"client_ip_to_uri_failed_requests_rate_limit,omitempty"`
	// Automatically track clients and classify them into 3 groups -
	// Good, Bad, Unknown. Clients are tracked based on their IP Address. Clients are added
	// to the Good group when the Avi Scan Detection system builds history of requests from them that complete
	// successfully. Clients are added to Unknown group when there is insufficient history about them.
	// Requests from such clients are rate limited to the rate specified in the RateProfile.
	// Finally, Clients with history of failed requests are added to Bad group and their requests
	// are rate limited with stricter thresholds than the Unknown Clients group. The Avi Scan Detection system
	// automatically tunes itself so that the Good, Bad,
	// and Unknown client IPs group membership changes dynamically with the changes in traffic patterns
	// through the ADC.
	// +optional
	ClientIPScannersRequestsRateLimit *RateProfile `json:"client_ip_scanners_requests_rate_limit,omitempty"`
	// Automatically track URIs and classify them into 3 groups -
	// Good, Bad, Unknown. URIs are added
	// to the Good group when the Avi Scan Detection system builds history of requests to URIs that complete
	// successfully. URIs are added to Unknown group when there is insufficient history about them.
	// Requests for such URIs are rate limited to the rate specified in the RateProfile.
	// Finally, URIs with history of failed requests are added to Bad group and requests to them
	// are rate limited with stricter thresholds than the Unknown URIs group. The Avi Scan Detection system
	// automatically tunes itself so that the Good, Bad,
	// and Unknown URIs group membership changes dynamically with the changes in traffic patterns
	// through the ADC.
	// +optional
	URIScannersRequestsRateLimit *RateProfile `json:"uri_scanners_requests_rate_limit,omitempty"`
	// Rate Limit all HTTP requests from all client "
	// IP addresses that contain any single HTTP header value.
	HTTPHeaderRateLimits []RateProfile `json:"http_header_rate_limits,omitempty"`
	// Rate Limit all HTTP requests that map to any custom
	// +optional
	CustomRequestsRateLimit *RateProfile `json:"custom_requests_rate_limit,omitempty"`
}

type HTTPApplicationProfile struct {
	// Allows HTTP requests, not just TCP connections, to be load balanced across servers. Proxied TCP connections to servers may be reused by multiple clients to improve performance. Not compatible with Preserve Client IP.
	// +optional
	// +kubebuilder:default:=true
	ConnectionMultiplexingEnabled bool `json:"connection_multiplexing_enabled,omitempty"`
	// The client's original IP address is inserted into an HTTP request header sent to the server. Servers may use this address for logging or other purposes, rather than Avi's source NAT address used in the Avi to server IP connection.
	// +optional
	// +kubebuilder:default:=true
	XffEnabled bool `json:"xff_enabled,omitempty"`
	// Provide a custom name for the X-Forwarded-For header sent to the servers.
	// +optional
	// +kubebuilder:default:="X-Forwarded-For"
	XffAlternateName string `json:"xff_alternate_name,omitempty"`
	// Configure how incoming X-Forwarded-For headers from the client are handled.
	// +optional
	// +kubebuilder:default:="REPLACE_XFF_HEADERS"
	XffUpdate XFFUpdate `json:"xff_update,omitempty"`
	// Inserts HTTP Strict-Transport-Security header in the HTTPS response. HSTS can help mitigate man-in-the-middle attacks by telling browsers that support HSTS that they should only access this site via HTTPS.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	HstsEnabled bool `json:"hsts_enabled,omitempty"`
	// Number of days for which the client should regard this virtual service as a known HSTS host.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	// +kubebuilder:default:=365
	// TODO In Basic and ESSENTIAL allowed value is 365
	HstsMaxAge uint64 `json:"hsts_max_age,omitempty"`
	// Insert the 'includeSubdomains' directive in the HTTP Strict-Transport-Security header. Adding the includeSubdomains directive signals the User-Agent that the HSTS Policy applies to this HSTS Host as well as any subdomains of the host's domain name.
	// +optional
	// +kubebuilder:default:=true
	// TODO In Basic and ESSENTIAL allowed value and default valuse is false
	HstsSubdomainsEnabled bool `json:"hsts_subdomains_enabled,omitempty"`
	// Mark server cookies with the 'Secure' attribute. Client browsers will not send a cookie marked as secure over an unencrypted connection. If Avi is terminating SSL from clients and passing it as HTTP to the server, the server may return cookies without the secure flag set.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	SecureCookieEnabled bool `json:"secure_cookie_enabled,omitempty"`
	// Mark HTTP cookies as HTTPonly. This helps mitigate cross site scripting attacks as browsers will not allow these cookies to be read by third parties, such as javascript.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	HttponlyEnabled bool `json:"httponly_enabled,omitempty"`
	// Client requests received via HTTP will be redirected to HTTPS.
	// +optional
	// +kubebuilder:default:=false
	// TODO In ESSENTIAL allowed value is false
	HttpToHttps bool `json:"http_to_https,omitempty"`
	// When terminating client SSL sessions at Avi, servers may incorrectly send redirect to clients as HTTP. This option will rewrite the server's redirect responses for this virtual service from HTTP to HTTPS.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value and default valuse is false
	ServerSideRedirectToHttps bool `json:"server_side_redirect_to_https,omitempty"`
	// Insert an X-Forwarded-Proto header in the request sent to the server. When the client connects via SSL, Avi terminates the SSL, and then forwards the requests to the servers via HTTP, so the servers can determine the original protocol via this header. In this example, the value will be 'https'.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	XForwardedProtoEnabled bool `json:"x_forwarded_proto_enabled,omitempty"`
	// The max allowed length of time between a client establishing a TCP connection and Avi receives the first byte of the client's HTTP request.
	// +optional
	// +kubebuilder:default:=30000
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=100000000
	// (units) = MILLISECONDS,
	// TODO In Basic and ESSENTIAL allowed value is 30000
	PostAcceptTimeout int32 `json:"post_accept_timeout,omitempty"`
	// The maximum length of time allowed for a client to transmit an entire request header. This helps mitigate various forms of SlowLoris attacks.
	// +optional
	// +kubebuilder:default:=10000
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=100000000
	// (units) = MILLISECONDS
	// TODO In Basic and ESSENTIAL allowed value is 10000
	ClientHeaderTimeout int32 `json:"client_header_timeout,omitempty"`
	// The maximum length of time allowed between consecutive read operations for a client request body. The value '0' specifies no timeout. This setting generally impacts the length of time allowed for a client to send a POST.
	// +optional
	// +kubebuilder:default:=30000
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100000000
	// TODO In Basic allow_any = true and in ESSENTIAL allowed value is 30000
	// (units) = MILLISECONDS
	ClientBodyTimeout int32 `json:"client_body_timeout,omitempty"`
	// The max idle time allowed between HTTP requests over a Keep-alive connection.
	// +optional
	// +kubebuilder:default:=30000
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=100000000
	// (units) = MILLISECONDS
	// TODO In ESSENTIAL allowed value is 30000
	KeepaliveTimeout int32 `json:"keepalive_timeout,omitempty"`
	// Maximum size in Kbytes of a single HTTP header in the client request.
	// +optional
	// +kubebuilder:default:=12
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=64
	// (units) = KB,
	// TODO In ESSENTIAL allowed value is 12
	ClientMaxHeaderSize int32 `json:"client_max_header_size,omitempty"`
	// Maximum size in Kbytes of all the client HTTP request headers. This value can be overriden by client_max_header_size if that is larger.
	// +optional
	// +kubebuilder:default:=48
	// (units) = KB,
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=256
	ClientMaxRequestSize int32 `json:"client_max_request_size,omitempty"`
	// Maximum size for the client request body. This limits the size of the client data that can be uploaded/posted as part of a single HTTP Request. Default 0 => Unlimited.
	// +optional
	// (units) = KB,
	// +kubebuilder:default:=0
	ClientMaxBodySize int64 `json:"client_max_body_size,omitempty"`
	// HTTP Caching config to use with this HTTP Profile.
	// +optional
	// TODO In Basic and ESSENTIAL allow_none is true
	CacheConfig *HttpCacheConfig `json:"cache_config,omitempty"`
	// Set of match/action rules that govern what happens when the client certificate request is enabled
	// +optional
	// TODO In Basic and ESSENTIAL allow_none is true
	SSLClientCertificateAction *SSLClientCertificateAction `json:"ssl_client_certificate_action,omitempty"`
	// Specifies whether the client side verification is set to none, request or require.
	// +optional
	// +kubebuilder:default:="SSL_CLIENT_CERTIFICATE_NONE"
	// TODO In Basic and ESSENTIAL allowed value is SSL_CLIENT_CERTIFICATE_NONE,SSL_CLIENT_CERTIFICATE_REQUIRE
	SSLClientCertificateMode SSLClientCertificateMode `json:"ssl_client_certificate_mode,omitempty"`
	// Select the PKI profile to be associated with the Virtual Service. This profile defines the Certificate Authority and Revocation List.
	// +optional
	// +kubebuilder:validation:Reference:type=PKIProfile
	PkiProfileUuid string `json:"pki_profile_uuid,omitempty"`

	// Enable Websockets proxy for traffic from clients to the virtual service. Connections to this VS start in HTTP mode. If the client requests an Upgrade to Websockets, and the server responds back with success, then the connection is upgraded to WebSockets mode.
	// +optional
	// +kubebuilder:default:=true
	WebsocketsEnabled bool `json:"websockets_enabled,omitempty"`
	// Send HTTP 'Keep-Alive' header to the client. By default, the timeout specified in the 'Keep-Alive Timeout' field will be used unless the 'Use App Keepalive Timeout' flag is set, in which case the timeout sent by the application will be honored.
	// +optional
	// +kubebuilder:default:=false
	KeepaliveHeader bool `json:"keepalive_header,omitempty"`
	// Use 'Keep-Alive' header timeout sent by application instead of sending the HTTP Keep-Alive Timeout.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	UseAppKeepaliveTimeout bool `json:"use_app_keepalive_timeout,omitempty"`
	// Allow use of dot (.) in HTTP header names, for instance Header.app.special: PickAppVersionX.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	AllowDotsInHeaderName bool `json:"allow_dots_in_header_name,omitempty"`
	// Disable keep-alive client side connections for older browsers based off MS Internet Explorer 6.0 (MSIE6). For some applications, this might break NTLM authentication for older clients based off MSIE6. For such applications, set this option to false to allow keep-alive connections.
	// +optional
	// +kubebuilder:default:=true
	// TODO In Basic and ESSENTIAL allowed value is true
	DisableKeepalivePostsMsie6 bool `json:"disable_keepalive_posts_msie6,omitempty"`
	// Enable request body buffering for POST requests. If enabled, max buffer size is set to lower of 32M or the value (non-zero) configured in client_max_body_size.
	// +optional
	// +kubebuilder:default:=false
	EnableRequestBodyBuffering bool `json:"enable_request_body_buffering,omitempty"`
	// Enable support for fire and forget feature. If enabled, request from client is forwarded to server even if client prematurely closes the connection
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	EnableFireAndForget bool `json:"enable_fire_and_forget,omitempty"`
	// Maximum size in Kbytes of all the HTTP response headers.
	// +optional
	// +kubebuilder:default:=48
	// (units) = KB,
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=256
	// TODO In ESSENTIAL allowed value is 48
	MaxResponseHeadersSize int32 `json:"max_response_headers_size,omitempty"`
	// +optional
	// Enable HTTP2 for traffic from clients to the virtual service.
	// TODO In Basic and ESSENTIAL allow_none is true
	Http2Enabled bool `json:"http2_enabled,omitempty"`

	// Avi will respond with 100-Continue response if Expect: 100-Continue header received from client
	// +optional
	// +kubebuilder:default:=true
	RespondWith100Continue bool `json:"respond_with_100_continue,omitempty"`
	// Enable HTTP request body metrics. If enabled, requests from clients are parsed and relevant statistics about them are gathered. Currently, it processes HTTP POST requests with Content-Type application/x-www-form-urlencoded or multipart/form-data, and adds the number of detected parameters to the l7_client.http_params_count. This is an experimental feature and it may have performance impact. Use it when detailed information about the number of HTTP POST parameters is needed, e.g. for WAF sizing.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	EnableRequestBodyMetrics bool `json:"enable_request_body_metrics,omitempty"`
	// Forward the Connection: Close header coming from backend server to the client if connection-switching is enabled, i.e. front-end and backend connections are bound together
	// +optional
	// +kubebuilder:default:=true
	FwdCloseHdrForBoundConnections bool `json:"fwd_close_hdr_for_bound_connections,omitempty"`
	// The max number of HTTP requests that can be sent over a Keep-Alive connection. '0' means unlimited.
	// +optional
	// +kubebuilder:default:=100
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000000
	// TODO In Basic and ESSENTIAL allowed value is 100
	// (special_values) = "{\"0\": \"Unlimited requests on a connection\"}",
	MaxKeepaliveRequests int32 `json:"max_keepalive_requests,omitempty"`
	// Disable strict check between TLS servername and HTTP Host name
	// +optional
	// +kubebuilder:default:=false
	DisableSniHostnameCheck bool `json:"disable_sni_hostname_check,omitempty"`

	// If enabled, an HTTP request on an SSL port will result in connection close instead of a 400 response.
	// +optional
	// +kubebuilder:default:=false
	// TODO In Basic and ESSENTIAL allowed value is false
	ResetConnHttpOnSslPort bool `json:"reset_conn_http_on_ssl_port,omitempty"`
	// Size of HTTP buffer in kB.
	// +optional
	// +kubebuilder:default:=0
	// (units) = KB,
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=256
	// TODO In Basic and ESSENTIAL allowed value is 0
	//(special_values) =  "{\"0\": \"Auto compute the size of buffer\"}",
	HttpUpstreamBufferSize uint32 `json:"http_upstream_buffer_size,omitempty"`
	// Enable chunk body merge for chunked transfer encoding response.
	// +optional
	// +kubebuilder:default:=true
	EnableChunkMerge bool `json:"enable_chunk_merge,omitempty"`
	// Specifies the HTTP/2 specific application profile parameters.
	// +optional
	// TODO In ESSENTIAL allow_none is true
	HTTP2Profile *HTTP2ApplicationProfile `json:"http2_profile,omitempty"`
	// Detect NTLM apps based on the HTTP Response from the server. Once detected, connection multiplexing will be disabled for that connection.
	// +optional
	// +kubebuilder:default:=true
	// TODO In Basic allow_any = true
	DetectNtlmApp bool `json:"detect_ntlm_app,omitempty"`
	// Detect client IP from user specified header.
	// +optional
	// +kubebuilder:default:=false
	UseTrueClientIP bool `json:"use_true_client_ip,omitempty"`
	// Detect client IP from user specified header at the configured index in the specified direction.
	// +optional
	TrueClientIP *TrueClientIPConfig `json:"true_client_ip,omitempty"`
	// Pass through X-ACCEL headers.
	// +optional
	// +kubebuilder:default:=false
	PassThroughXAccelHeaders bool `json:"pass_through_x_accel_headers,omitempty"`
	// If enabled, the client's TLS fingerprint will be collected and included in the Application Log. For Virtual Services with Bot Detection enabled, TLS fingerprints are always computed if 'use_tls_fingerprint' is enabled in the Bot Detection Policy's User-Agent detection component.
	// +optional
	// +kubebuilder:default:=false
	CollectClientTlsFingerprint bool `json:"collect_client_tls_fingerprint,omitempty"`
	// Maximum number of headers allowed in HTTP request and response.
	// +optional
	// +kubebuilder:default:=256
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4096
	// TODO In Basic and ESSENTIAL allowed value is 0 and default_value is 0
	// (special_values) =  "{\"0\": \"unlimited headers in request and response\"}",
	MaxHeaderCount int32 `json:"max_header_count,omitempty"`
	// HTTP session configuration.
	// +optional
	HTTPSessionConfig *HTTPSessionConfig `json:"session_config,omitempty"`
	// Close server-side connection when an error response is received.
	// +optional
	// +kubebuilder:default:=false
	CloseServerSideConnectionOnError bool `json:"close_server_side_connection_on_error,omitempty"`
}

type TrueClientIPConfig struct {
	// +kubebuilder:validation:MaxItems=1
	// TODO May be add validation webhook for checking stringlength to be Max 128
	// HTTP Headers to derive client IP from. If none specified and use_true_client_ip is set to true, it will use X-Forwarded-For header, if present.
	// +kubebuilder:default:={"X-Forwarded-For"}
	Headers []string `json:"headers,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default:=1
	// +optional
	// Position in the configured direction, in the specified header's value, to be used to set true client IP. If the value is greater than the number of IP addresses in the header, then the last IP address in the configured direction in the header will be used.
	IndexInHeader uint32 `json:"index_in_header,omitempty"`

	// +kubebuilder:default:="LEFT"
	// +optional
	// Denotes the end from which to count the IPs in the specified header value.
	Direction TrueClientIPIndexDirection `json:"direction,omitempty"`
}

type HTTPSessionConfig struct {
	// HTTP session cookie name to use.
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=`^[A-Za-z_]+$`
	// +kubebuilder:default:="albsessid"
	// +optional
	SessionCookieName string `json:"session_cookie_name,omitempty"`

	// HTTP session cookie SameSite attribute.
	// +kubebuilder:default:="SAMESITE_LAX"
	// +optional
	SessionCookieSameSite HTTPCookieSameSite `json:"session_cookie_samesite,omitempty"`

	// If set, HTTP session cookie will use 'Secure' attribute.
	// +kubebuilder:default:=true
	// +optional
	SessionCookieSecure bool `json:"session_cookie_secure,omitempty"`

	// If set, HTTP session cookie will use 'HttpOnly' attribute.
	// +kubebuilder:default:=true
	// +optional
	SessionCookieHttponly bool `json:"session_cookie_httponly,omitempty"`

	// Maximum allowed time between creating a session and the client coming back. Value in seconds.
	// +kubebuilder:validation:Minimum=120
	// +kubebuilder:validation:Maximum=3600
	// +kubebuilder:default:=300
	// +optional
	// Units=SEC
	SessionEstablishmentTimeout uint32 `json:"session_establishment_timeout,omitempty"`

	// Maximum allowed time to expire the session after establishment on client inactivity. Value in seconds.
	// +kubebuilder:validation:Minimum=120
	// +kubebuilder:validation:Maximum=604800
	// +kubebuilder:default:=1800
	// +optional
	// Units=SEC
	SessionIdleTimeout uint32 `json:"session_idle_timeout,omitempty"`

	// Maximum allowed time to expire the session, even if it is still active. Value in seconds.
	// +kubebuilder:validation:Minimum=120
	// +kubebuilder:validation:Maximum=604800
	// +kubebuilder:default:=28800
	// +optional
	// +kubebuilder:validation:Units=SEC
	SessionMaximumTimeout uint32 `json:"session_maximum_timeout,omitempty"`
}

type HTTP2ApplicationProfile struct {
	// Maximum number of control frames that client can send over an HTTP/2 connection. '0' means unlimited.
	// +kubebuilder:default:=1000
	// +kubebuilder:validation:Minimum=0
	// +optional
	// +kubebuilder:validation:Maximum=10000
	// (special_values) = "{\"0\": \"Unlimited control frames on a client side HTTP/2 connection\"}",
	MaxHTTP2ControlFramesPerConnection uint32 `json:"max_http2_control_frames_per_connection,omitempty"`

	// Maximum number of frames that can be queued waiting to be sent over a client side HTTP/2 connection at any given time. '0' means unlimited.
	// +kubebuilder:default:=1000
	// +kubebuilder:validation:Minimum=0
	// +optional
	// +kubebuilder:validation:Maximum=10000
	// (special_values) = "{\"0\": \"Unlimited frames can be queued on a client side HTTP/2 connection\"}",
	MaxHTTP2QueuedFramesToClientPerConnection uint32 `json:"max_http2_queued_frames_to_client_per_connection,omitempty"`

	// Maximum number of empty data frames that client can send over an HTTP/2 connection. '0' means unlimited.
	// +kubebuilder:default:=1000
	// +kubebuilder:validation:Minimum=0
	// +optional
	// +kubebuilder:validation:Maximum=10000
	// (special_values) = "{\"0\": \"Unlimited empty data frames over a client side HTTP/2 connection\"}",
	MaxHTTP2EmptyDataFramesPerConnection uint32 `json:"max_http2_empty_data_frames_per_connection,omitempty"`

	// Maximum number of concurrent streams over a client side HTTP/2 connection.
	// +kubebuilder:default:=128
	// +kubebuilder:validation:Minimum=1
	// +optional
	// +kubebuilder:validation:Maximum=256
	MaxHTTP2ConcurrentStreamsPerConnection uint32 `json:"max_http2_concurrent_streams_per_connection,omitempty"`

	// Maximum number of requests over a client side HTTP/2 connection.
	// +kubebuilder:default:=1000
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	// +optional
	// (special_values) = "{\"0\": \"Unlimited requests on a client side HTTP/2 connection\"}",
	MaxHTTP2RequestsPerConnection uint32 `json:"max_http2_requests_per_connection,omitempty"`

	// Maximum size in bytes of the compressed request header field. The limit applies equally to both name and value.
	// +kubebuilder:default:=4096
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8192
	// +optional
	// (units) = BYTES,
	MaxHTTP2HeaderFieldSize uint32 `json:"max_http2_header_field_size,omitempty"`

	// The initial flow control window size in KB for HTTP/2 streams.
	// +kubebuilder:default:=64
	// +kubebuilder:validation:Minimum=64
	// +kubebuilder:validation:Maximum=32768
	// +optional
	// (units) = KB,
	HTTP2InitialWindowSize uint32 `json:"http2_initial_window_size,omitempty"`

	// Enables automatic conversion of preload links specified in the "Link" response header fields into Server push requests.
	// +kubebuilder:default:=false
	EnableHTTP2ServerPush bool `json:"enable_http2_server_push,omitempty"`

	// Maximum number of concurrent push streams over a client side HTTP/2 connection.
	// +kubebuilder:default:=10
	// +kubebuilder:validation:Minimum=1
	// +optional
	// +kubebuilder:validation:Maximum=256
	MaxHTTP2ConcurrentPushesPerConnection uint32 `json:"max_http2_concurrent_pushes_per_connection,omitempty"`
}

type RoleFilterMatchLabel struct {
	// Key for filter match.
	// +kubebuilder:validation:Required
	// option (key) = "key";
	Key string `json:"key"`

	// Values for filter match. Multiple values will be evaluated as OR. Example: key = value1 OR key = value2. Behavior for match is key = * if this field is empty.
	Values []string `json:"values,omitempty"`
}

// ApplicationProfileSpec defines the desired state of ApplicationProfile
// Can be created in BASIC and ESSENTIAL license tiers
type ApplicationProfileSpec struct {
	// Type specifies which application layer proxy is enabled for the virtual service.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="applicatio_profile_type is immutable"
	// TODO in BASIC tier allowed_values are APPLICATION_PROFILE_TYPE_L4,APPLICATION_PROFILE_TYPE_HTTP and in ESSENTIALS toer allowed_value is APPLICATION_PROFILE_TYPE_L4
	Type ApplicationProfileType `json:"type,omitempty"`

	// HTTPProfile specifies the HTTP application proxy profile parameters.
	// +optional
	// TODO in ESSENTIALS allow_none is true
	HTTPProfile *HTTPApplicationProfile `json:"http_profile,omitempty"`

	// PreserveClientIP specifies if client IP needs to be preserved for backend connection. Not compatible with Connection Multiplexing.
	// +optional
	// +kubebuilder:default:=false
	PreserveClientIP bool `json:"preserve_client_ip,omitempty"`

	// PreserveClientPort specifies if we need to preserve client port while preserving client IP for backend connections.
	// +optional
	PreserveClientPort bool `json:"preserve_client_port,omitempty"`

	// PreserveDestIpPort specifies if destination IP and port needs to be preserved for backend connection.
	// +optional
	// +kubebuilder:default:=false
	// TODO in BASIC and ESSENTIALS allowed_value is false
	PreserveDestIpPort bool `json:"preserve_dest_ip_port,omitempty"`

	// List of labels to be used for granular RBAC.
	// TODO in BASIC and ESSENTIALS allow_any is true
	Markers []RoleFilterMatchLabel `json:"markers,omitempty"`

	// Description is a description of the application profile.
	// +optional
	Description string `json:"description,omitempty"`

	// AppServiceType specifies app service type for an application.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="app_service_type is immutable"
	AppServiceType AppServiceType `json:"app_service_type,omitempty"`
}

// ApplicationProfileStatus defines the observed state of ApplicationProfile
type ApplicationProfileStatus struct {
	// Status of the application profile
	Status string `json:"status,omitempty"`
	// Error if any error was encountered
	Error string `json:"error"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=applicationprofiles,scope=Namespaced
// +kubebuilder:subresource:status
// ApplicationProfile is the Schema for the applicationprofiles API
type ApplicationProfile struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of ApplicationProfile
	// +optional
	Spec ApplicationProfileSpec `json:"spec,omitempty"`

	// Status defines the observed state of ApplicationProfile
	// +optional
	Status ApplicationProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ApplicationProfileList contains a list of ApplicationProfile.
type ApplicationProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApplicationProfile{}, &ApplicationProfileList{})
}

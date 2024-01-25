// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPApplicationProfile HTTP application profile
// swagger:model HTTPApplicationProfile
type HTTPApplicationProfile struct {

	// Allow use of dot (.) in HTTP header names, for instance Header.app.special  PickAppVersionX. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AllowDotsInHeaderName *bool `json:"allow_dots_in_header_name,omitempty"`

	// HTTP Caching config to use with this HTTP Profile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CacheConfig *HTTPCacheConfig `json:"cache_config,omitempty"`

	// The maximum length of time allowed between consecutive read operations for a client request body. The value '0' specifies no timeout. This setting generally impacts the length of time allowed for a client to send a POST. Allowed values are 0-100000000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 30000), Basic edition with any value, Enterprise with Cloud Services edition.
	ClientBodyTimeout *int32 `json:"client_body_timeout,omitempty"`

	// The maximum length of time allowed for a client to transmit an entire request header. This helps mitigate various forms of SlowLoris attacks. Allowed values are 10-100000000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 10000), Basic edition(Allowed values- 10000), Enterprise with Cloud Services edition.
	ClientHeaderTimeout *int32 `json:"client_header_timeout,omitempty"`

	// Maximum size for the client request body.  This limits the size of the client data that can be uploaded/posted as part of a single HTTP Request.  Default 0 => Unlimited. Unit is KB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientMaxBodySize *int64 `json:"client_max_body_size,omitempty"`

	// Maximum size in Kbytes of a single HTTP header in the client request. Allowed values are 1-64. Unit is KB. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 12), Basic, Enterprise with Cloud Services edition.
	ClientMaxHeaderSize *int32 `json:"client_max_header_size,omitempty"`

	// Maximum size in Kbytes of all the client HTTP request headers.This value can be overriden by client_max_header_size if that is larger. Allowed values are 1-256. Unit is KB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientMaxRequestSize *int32 `json:"client_max_request_size,omitempty"`

	// Close server-side connection when an error response is received. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloseServerSideConnectionOnError *bool `json:"close_server_side_connection_on_error,omitempty"`

	// If enabled, the client's TLS fingerprint will be collected and included in the Application Log. For Virtual Services with Bot Detection enabled, TLS fingerprints are always computed if 'use_tls_fingerprint' is enabled in the Bot Detection Policy's User-Agent detection component. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CollectClientTLSFingerprint *bool `json:"collect_client_tls_fingerprint,omitempty"`

	// HTTP Compression settings to use with this HTTP Profile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CompressionProfile *CompressionProfile `json:"compression_profile,omitempty"`

	// Allows HTTP requests, not just TCP connections, to be load balanced across servers.  Proxied TCP connections to servers may be reused by multiple clients to improve performance. Not compatible with Preserve Client IP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnectionMultiplexingEnabled *bool `json:"connection_multiplexing_enabled,omitempty"`

	// Detect NTLM apps based on the HTTP Response from the server. Once detected, connection multiplexing will be disabled for that connection. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	DetectNtlmApp *bool `json:"detect_ntlm_app,omitempty"`

	// Disable keep-alive client side connections for older browsers based off MS Internet Explorer 6.0 (MSIE6). For some applications, this might break NTLM authentication for older clients based off MSIE6. For such applications, set this option to false to allow keep-alive connections. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	DisableKeepalivePostsMsie6 *bool `json:"disable_keepalive_posts_msie6,omitempty"`

	// Disable strict check between TLS servername and HTTP Host name. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableSniHostnameCheck *bool `json:"disable_sni_hostname_check,omitempty"`

	// Enable chunk body merge for chunked transfer encoding response. Field introduced in 18.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableChunkMerge *bool `json:"enable_chunk_merge,omitempty"`

	// Enable support for fire and forget feature. If enabled, request from client is forwarded to server even if client prematurely closes the connection. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EnableFireAndForget *bool `json:"enable_fire_and_forget,omitempty"`

	// Enable request body buffering for POST requests. If enabled, max buffer size is set to lower of 32M or the value (non-zero) configured in client_max_body_size. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableRequestBodyBuffering *bool `json:"enable_request_body_buffering,omitempty"`

	// Enable HTTP request body metrics. If enabled, requests from clients are parsed and relevant statistics about them are gathered. Currently, it processes HTTP POST requests with Content-Type application/x-www-form-urlencoded or multipart/form-data, and adds the number of detected parameters to the l7_client.http_params_count. This is an experimental feature and it may have performance impact. Use it when detailed information about the number of HTTP POST parameters is needed, e.g. for WAF sizing. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EnableRequestBodyMetrics *bool `json:"enable_request_body_metrics,omitempty"`

	// Forward the Connection  Close header coming from backend server to the client if connection-switching is enabled, i.e. front-end and backend connections are bound together. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FwdCloseHdrForBoundConnections *bool `json:"fwd_close_hdr_for_bound_connections,omitempty"`

	// Inserts HTTP Strict-Transport-Security header in the HTTPS response.  HSTS can help mitigate man-in-the-middle attacks by telling browsers that support HSTS that they should only access this site via HTTPS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	HstsEnabled *bool `json:"hsts_enabled,omitempty"`

	// Number of days for which the client should regard this virtual service as a known HSTS host. Allowed values are 0-10000. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 365), Basic edition(Allowed values- 365), Enterprise with Cloud Services edition.
	HstsMaxAge *uint64 `json:"hsts_max_age,omitempty"`

	// Insert the 'includeSubdomains' directive in the HTTP Strict-Transport-Security header. Adding the includeSubdomains directive signals the User-Agent that the HSTS Policy applies to this HSTS Host as well as any subdomains of the host's domain name. Field introduced in 17.2.13, 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Essentials edition is false, Basic edition is false, Enterprise is True.
	HstsSubdomainsEnabled *bool `json:"hsts_subdomains_enabled,omitempty"`

	// Specifies the HTTP/2 specific application profile parameters. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	Http2Profile *Http2ApplicationProfile `json:"http2_profile,omitempty"`

	// Client requests received via HTTP will be redirected to HTTPS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic, Enterprise with Cloud Services edition.
	HTTPToHTTPS *bool `json:"http_to_https,omitempty"`

	// Size of HTTP buffer in kB. Allowed values are 1-256. Special values are 0- Auto compute the size of buffer. Field introduced in 20.1.1. Unit is KB. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	HTTPUpstreamBufferSize uint32 `json:"http_upstream_buffer_size,omitempty"`

	// Mark HTTP cookies as HTTPonly.  This helps mitigate cross site scripting attacks as browsers will not allow these cookies to be read by third parties, such as javascript. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	HttponlyEnabled *bool `json:"httponly_enabled,omitempty"`

	// Send HTTP 'Keep-Alive' header to the client. By default, the timeout specified in the 'Keep-Alive Timeout' field will be used unless the 'Use App Keepalive Timeout' flag is set, in which case the timeout sent by the application will be honored. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeepaliveHeader *bool `json:"keepalive_header,omitempty"`

	// The max idle time allowed between HTTP requests over a Keep-alive connection. Allowed values are 10-100000000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 30000), Basic, Enterprise with Cloud Services edition.
	KeepaliveTimeout *int32 `json:"keepalive_timeout,omitempty"`

	// Maximum bad requests per second per client IP. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxBadRpsCip uint32 `json:"max_bad_rps_cip,omitempty"`

	// Maximum bad requests per second per client IP and URI. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxBadRpsCipURI uint32 `json:"max_bad_rps_cip_uri,omitempty"`

	// Maximum bad requests per second per URI. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxBadRpsURI uint32 `json:"max_bad_rps_uri,omitempty"`

	// Maximum number of headers allowed in HTTP request and response. Allowed values are 0-4096. Special values are 0- unlimited headers in request and response. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition. Special default for Essentials edition is 0, Basic edition is 0, Enterprise is 256.
	MaxHeaderCount *int32 `json:"max_header_count,omitempty"`

	// The max number of HTTP requests that can be sent over a Keep-Alive connection. '0' means unlimited. Allowed values are 0-1000000. Special values are 0- Unlimited requests on a connection. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 100), Basic edition(Allowed values- 100), Enterprise with Cloud Services edition.
	MaxKeepaliveRequests *int32 `json:"max_keepalive_requests,omitempty"`

	// Maximum size in Kbytes of all the HTTP response headers. Allowed values are 1-256. Unit is KB. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 48), Basic, Enterprise with Cloud Services edition.
	MaxResponseHeadersSize *int32 `json:"max_response_headers_size,omitempty"`

	// Maximum requests per second per client IP. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRpsCip uint32 `json:"max_rps_cip,omitempty"`

	// Maximum requests per second per client IP and URI. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRpsCipURI uint32 `json:"max_rps_cip_uri,omitempty"`

	// Maximum unknown client IPs per second. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRpsUnknownCip uint32 `json:"max_rps_unknown_cip,omitempty"`

	// Maximum unknown URIs per second. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRpsUnknownURI uint32 `json:"max_rps_unknown_uri,omitempty"`

	// Maximum requests per second per URI. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRpsURI uint32 `json:"max_rps_uri,omitempty"`

	// Pass through X-ACCEL headers. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PassThroughXAccelHeaders *bool `json:"pass_through_x_accel_headers,omitempty"`

	// Select the PKI profile to be associated with the Virtual Service. This profile defines the Certificate Authority and Revocation List. It is a reference to an object of type PKIProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// The max allowed length of time between a client establishing a TCP connection and Avi receives the first byte of the client's HTTP request. Allowed values are 10-100000000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 30000), Basic edition(Allowed values- 30000), Enterprise with Cloud Services edition.
	PostAcceptTimeout *int32 `json:"post_accept_timeout,omitempty"`

	// If enabled, an HTTP request on an SSL port will result in connection close instead of a 400 response. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ResetConnHTTPOnSslPort *bool `json:"reset_conn_http_on_ssl_port,omitempty"`

	// Avi will respond with 100-Continue response if Expect  100-Continue header received from client. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RespondWith100Continue *bool `json:"respond_with_100_continue,omitempty"`

	// Mark server cookies with the 'Secure' attribute.  Client browsers will not send a cookie marked as secure over an unencrypted connection.  If Avi is terminating SSL from clients and passing it as HTTP to the server, the server may return cookies without the secure flag set. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	SecureCookieEnabled *bool `json:"secure_cookie_enabled,omitempty"`

	// When terminating client SSL sessions at Avi, servers may incorrectly send redirect to clients as HTTP.  This option will rewrite the server's redirect responses for this virtual service from HTTP to HTTPS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ServerSideRedirectToHTTPS *bool `json:"server_side_redirect_to_https,omitempty"`

	// HTTP session configuration. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionConfig *HttpsessionConfig `json:"session_config,omitempty"`

	// Set of match/action rules that govern what happens when the client certificate request is enabled. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslClientCertificateAction *SSLClientCertificateAction `json:"ssl_client_certificate_action,omitempty"`

	// Specifies whether the client side verification is set to none, request or require. Enum options - SSL_CLIENT_CERTIFICATE_NONE, SSL_CLIENT_CERTIFICATE_REQUEST, SSL_CLIENT_CERTIFICATE_REQUIRE. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- SSL_CLIENT_CERTIFICATE_NONE,SSL_CLIENT_CERTIFICATE_REQUIRE), Basic edition(Allowed values- SSL_CLIENT_CERTIFICATE_NONE,SSL_CLIENT_CERTIFICATE_REQUIRE), Enterprise with Cloud Services edition.
	SslClientCertificateMode *string `json:"ssl_client_certificate_mode,omitempty"`

	// Detect client IP from user specified header at the configured index in the specified direction. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TrueClientIP *TrueClientIPConfig `json:"true_client_ip,omitempty"`

	// Use 'Keep-Alive' header timeout sent by application instead of sending the HTTP Keep-Alive Timeout. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	UseAppKeepaliveTimeout *bool `json:"use_app_keepalive_timeout,omitempty"`

	// Detect client IP from user specified header. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UseTrueClientIP *bool `json:"use_true_client_ip,omitempty"`

	// Enable Websockets proxy for traffic from clients to the virtual service. Connections to this VS start in HTTP mode. If the client requests an Upgrade to Websockets, and the server responds back with success, then the connection is upgraded to WebSockets mode. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WebsocketsEnabled *bool `json:"websockets_enabled,omitempty"`

	// Insert an X-Forwarded-Proto header in the request sent to the server.  When the client connects via SSL, Avi terminates the SSL, and then forwards the requests to the servers via HTTP, so the servers can determine the original protocol via this header.  In this example, the value will be 'https'. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	XForwardedProtoEnabled *bool `json:"x_forwarded_proto_enabled,omitempty"`

	// Provide a custom name for the X-Forwarded-For header sent to the servers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	XffAlternateName *string `json:"xff_alternate_name,omitempty"`

	// The client's original IP address is inserted into an HTTP request header sent to the server.  Servers may use this address for logging or other purposes, rather than Avi's source NAT address used in the Avi to server IP connection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	XffEnabled *bool `json:"xff_enabled,omitempty"`

	// Configure how incoming X-Forwarded-For headers from the client are handled. Enum options - REPLACE_XFF_HEADERS, APPEND_TO_THE_XFF_HEADER, ADD_NEW_XFF_HEADER. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	XffUpdate *string `json:"xff_update,omitempty"`
}

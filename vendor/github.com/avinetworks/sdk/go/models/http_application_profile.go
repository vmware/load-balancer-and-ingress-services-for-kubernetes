package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPApplicationProfile HTTP application profile
// swagger:model HTTPApplicationProfile
type HTTPApplicationProfile struct {

	// Allow use of dot (.) in HTTP header names, for instance Header.app.special  PickAppVersionX.
	AllowDotsInHeaderName *bool `json:"allow_dots_in_header_name,omitempty"`

	// HTTP Caching config to use with this HTTP Profile.
	CacheConfig *HTTPCacheConfig `json:"cache_config,omitempty"`

	// The maximum length of time allowed between consecutive read operations for a client request body. The value '0' specifies no timeout. This setting generally impacts the length of time allowed for a client to send a POST. Allowed values are 0-100000000.
	ClientBodyTimeout *int32 `json:"client_body_timeout,omitempty"`

	// The maximum length of time allowed for a client to transmit an entire request header. This helps mitigate various forms of SlowLoris attacks. Allowed values are 10-100000000.
	ClientHeaderTimeout *int32 `json:"client_header_timeout,omitempty"`

	// Maximum size for the client request body.  This limits the size of the client data that can be uploaded/posted as part of a single HTTP Request.  Default 0 => Unlimited.
	ClientMaxBodySize *int64 `json:"client_max_body_size,omitempty"`

	// Maximum size in Kbytes of a single HTTP header in the client request. Allowed values are 1-64.
	ClientMaxHeaderSize *int32 `json:"client_max_header_size,omitempty"`

	// Maximum size in Kbytes of all the client HTTP request headers. Allowed values are 1-256.
	ClientMaxRequestSize *int32 `json:"client_max_request_size,omitempty"`

	// HTTP Compression settings to use with this HTTP Profile.
	CompressionProfile *CompressionProfile `json:"compression_profile,omitempty"`

	// Allows HTTP requests, not just TCP connections, to be load balanced across servers.  Proxied TCP connections to servers may be reused by multiple clients to improve performance. Not compatible with Preserve Client IP.
	ConnectionMultiplexingEnabled *bool `json:"connection_multiplexing_enabled,omitempty"`

	// Disable keep-alive client side connections for older browsers based off MS Internet Explorer 6.0 (MSIE6). For some applications, this might break NTLM authentication for older clients based off MSIE6. For such applications, set this option to false to allow keep-alive connections.
	DisableKeepalivePostsMsie6 *bool `json:"disable_keepalive_posts_msie6,omitempty"`

	// Enable support for fire and forget feature. If enabled, request from client is forwarded to server even if client prematurely closes the connection. Field introduced in 17.2.4.
	EnableFireAndForget *bool `json:"enable_fire_and_forget,omitempty"`

	// Enable request body buffering for POST requests. If enabled, max buffer size is set to lower of 32M or the value (non-zero) configured in client_max_body_size.
	EnableRequestBodyBuffering *bool `json:"enable_request_body_buffering,omitempty"`

	// Enable HTTP request body metrics. If enabled, requests from clients are parsed and relevant statistics about them are gathered. Currently, it processes HTTP POST requests with Content-Type application/x-www-form-urlencoded or multipart/form-data, and adds the number of detected parameters to the l7_client.http_params_count. This is an experimental feature and it may have performance impact. Use it when detailed information about the number of HTTP POST parameters is needed, e.g. for WAF sizing. Field introduced in 18.1.5, 18.2.1.
	EnableRequestBodyMetrics *bool `json:"enable_request_body_metrics,omitempty"`

	// Inserts HTTP Strict-Transport-Security header in the HTTPS response.  HSTS can help mitigate man-in-the-middle attacks by telling browsers that support HSTS that they should only access this site via HTTPS.
	HstsEnabled *bool `json:"hsts_enabled,omitempty"`

	// Number of days for which the client should regard this virtual service as a known HSTS host. Allowed values are 0-10000.
	HstsMaxAge *int64 `json:"hsts_max_age,omitempty"`

	// Insert the 'includeSubdomains' directive in the HTTP Strict-Transport-Security header. Adding the includeSubdomains directive signals the User-Agent that the HSTS Policy applies to this HSTS Host as well as any subdomains of the host's domain name. Field introduced in 17.2.13, 18.1.4, 18.2.1.
	HstsSubdomainsEnabled *bool `json:"hsts_subdomains_enabled,omitempty"`

	// Enable HTTP2 for traffic from clients to the virtual service.  . Field introduced in 18.1.1.
	Http2Enabled *bool `json:"http2_enabled,omitempty"`

	// Client requests received via HTTP will be redirected to HTTPS.
	HTTPToHTTPS *bool `json:"http_to_https,omitempty"`

	// Mark HTTP cookies as HTTPonly.  This helps mitigate cross site scripting attacks as browsers will not allow these cookies to be read by third parties, such as javascript.
	HttponlyEnabled *bool `json:"httponly_enabled,omitempty"`

	// Send HTTP 'Keep-Alive' header to the client. By default, the timeout specified in the 'Keep-Alive Timeout' field will be used unless the 'Use App Keepalive Timeout' flag is set, in which case the timeout sent by the application will be honored.
	KeepaliveHeader *bool `json:"keepalive_header,omitempty"`

	// The max idle time allowed between HTTP requests over a Keep-alive connection. Allowed values are 10-100000000.
	KeepaliveTimeout *int32 `json:"keepalive_timeout,omitempty"`

	// Maximum bad requests per second per client IP. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxBadRpsCip *int32 `json:"max_bad_rps_cip,omitempty"`

	// Maximum bad requests per second per client IP and URI. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxBadRpsCipURI *int32 `json:"max_bad_rps_cip_uri,omitempty"`

	// Maximum bad requests per second per URI. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxBadRpsURI *int32 `json:"max_bad_rps_uri,omitempty"`

	// Maximum size in Kbytes of all the HTTP response headers. Allowed values are 1-256.
	MaxResponseHeadersSize *int32 `json:"max_response_headers_size,omitempty"`

	// Maximum requests per second per client IP. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxRpsCip *int32 `json:"max_rps_cip,omitempty"`

	// Maximum requests per second per client IP and URI. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxRpsCipURI *int32 `json:"max_rps_cip_uri,omitempty"`

	// Maximum unknown client IPs per second. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxRpsUnknownCip *int32 `json:"max_rps_unknown_cip,omitempty"`

	// Maximum unknown URIs per second. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxRpsUnknownURI *int32 `json:"max_rps_unknown_uri,omitempty"`

	// Maximum requests per second per URI. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxRpsURI *int32 `json:"max_rps_uri,omitempty"`

	// Select the PKI profile to be associated with the Virtual Service. This profile defines the Certificate Authority and Revocation List. It is a reference to an object of type PKIProfile.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// The max allowed length of time between a client establishing a TCP connection until Avi receives the first byte of the client's HTTP request. Allowed values are 10-100000000.
	PostAcceptTimeout *int32 `json:"post_accept_timeout,omitempty"`

	// Avi will respond with 100-Continue response if Expect  100-Continue header received from client. Field introduced in 17.2.8.
	RespondWith100Continue *bool `json:"respond_with_100_continue,omitempty"`

	// Mark server cookies with the 'Secure' attribute.  Client browsers will not send a cookie marked as secure over an unencrypted connection.  If Avi is terminating SSL from clients and passing it as HTTP to the server, the server may return cookies without the secure flag set.
	SecureCookieEnabled *bool `json:"secure_cookie_enabled,omitempty"`

	// When terminating client SSL sessions at Avi, servers may incorrectly send redirect to clients as HTTP.  This option will rewrite the server's redirect responses for this virtual service from HTTP to HTTPS.
	ServerSideRedirectToHTTPS *bool `json:"server_side_redirect_to_https,omitempty"`

	// Enable SPDY proxy for traffic from clients to the virtual service.  SPDY requires SSL from the clients to Avi.  Avi ADC will proxy the SPDY protocol, and forward requests to servers as HTTP 1.1. .
	SpdyEnabled *bool `json:"spdy_enabled,omitempty"`

	// Enable fwd proxy mode with SPDY. This makes the Proxy combine the  host and  uri spdy headers to create a fwd-proxy style request URI.
	SpdyFwdProxyMode *bool `json:"spdy_fwd_proxy_mode,omitempty"`

	// Set of match/action rules that govern what happens when the client certificate request is enabled.
	SslClientCertificateAction *SSLClientCertificateAction `json:"ssl_client_certificate_action,omitempty"`

	// Specifies whether the client side verification is set to none, request or require. Enum options - SSL_CLIENT_CERTIFICATE_NONE, SSL_CLIENT_CERTIFICATE_REQUEST, SSL_CLIENT_CERTIFICATE_REQUIRE.
	SslClientCertificateMode *string `json:"ssl_client_certificate_mode,omitempty"`

	// Enable common settings to increase the level of security for  virtual services running HTTP and HTTPS.  For sites that are  HTTP only, these settings will have no effect.
	SslEverywhereEnabled *bool `json:"ssl_everywhere_enabled,omitempty"`

	// Use 'Keep-Alive' header timeout sent by application instead of sending the HTTP Keep-Alive Timeout.
	UseAppKeepaliveTimeout *bool `json:"use_app_keepalive_timeout,omitempty"`

	// Enable Websockets proxy for traffic from clients to the virtual service. Connections to this VS start in HTTP mode. If the client requests an Upgrade to Websockets, and the server responds back with success, then the connection is upgraded to WebSockets mode. .
	WebsocketsEnabled *bool `json:"websockets_enabled,omitempty"`

	// Insert an X-Forwarded-Proto header in the request sent to the server.  When the client connects via SSL, Avi terminates the SSL, and then forwards the requests to the servers via HTTP, so the servers can determine the original protocol via this header.  In this example, the value will be 'https'.
	XForwardedProtoEnabled *bool `json:"x_forwarded_proto_enabled,omitempty"`

	// Provide a custom name for the X-Forwarded-For header sent to the servers.
	XffAlternateName *string `json:"xff_alternate_name,omitempty"`

	// The client's original IP address is inserted into an HTTP request header sent to the server.  Servers may use this address for logging or other purposes, rather than Avi's source NAT address used in the Avi to server IP connection.
	XffEnabled *bool `json:"xff_enabled,omitempty"`
}

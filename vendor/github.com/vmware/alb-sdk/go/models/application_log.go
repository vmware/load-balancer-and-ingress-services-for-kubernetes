// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ApplicationLog application log
// swagger:model ApplicationLog
type ApplicationLog struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Adf *bool `json:"adf"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllRequestHeaders *string `json:"all_request_headers,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllResponseHeaders *string `json:"all_response_headers,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppResponseTime uint64 `json:"app_response_time,omitempty"`

	// Set the Session Authentication Status. Enum options - AUTH_STATUS_NO_AUTHENTICATION, AUTH_STATUS_AUTHENTICATION_SUCCESS, AUTH_STATUS_AUTHENTICATION_FAILURE, AUTH_STATUS_UNAUTHORIZED, AUTH_STATUS_AUTHENTICATED_REQUEST, AUTH_STATUS_AUTHZ_FAILED. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AuthStatus *string `json:"auth_status,omitempty"`

	// Average packet processing latency for the backend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AvgIngressLatencyBe uint32 `json:"avg_ingress_latency_be,omitempty"`

	// Average packet processing latency for the frontend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AvgIngressLatencyFe uint32 `json:"avg_ingress_latency_fe,omitempty"`

	//  Enum options - NOT_UPDATED, BY_CONTENT_REWRITE_PROFILE, BY_DATA_SCRIPT. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BodyUpdated *string `json:"body_updated,omitempty"`

	// Logs related to Bot detection. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BotManagementLog *BotManagementLog `json:"bot_management_log,omitempty"`

	// Cache fetch and store is disabled by the Datascript policies. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CacheDisabledByDs *bool `json:"cache_disabled_by_ds,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CacheHit *bool `json:"cache_hit,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Cacheable *bool `json:"cacheable,omitempty"`

	// Byte stream of client cipher list sent on SSL_R_NO_SHARED_CIPHER error.This byte stream is used to generate client_cipher_list. Field introduced in 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	CipherBytes *string `json:"cipher_bytes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientBrowser *string `json:"client_browser,omitempty"`

	// List of ciphers sent by client in TLS Client Hello. This field is only generated when TLS handshake fails due to no shared cipher. Field introduced in 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientCipherList *SSLCipherList `json:"client_cipher_list,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientDestPort *uint32 `json:"client_dest_port"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientDevice *string `json:"client_device,omitempty"`

	// The fingerprints for this client. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientFingerprints *ClientFingerprints `json:"client_fingerprints,omitempty"`

	//  Enum options - INSIGHTS_DISABLED, NO_INSIGHTS_NOT_SAMPLED_COUNT, NO_INSIGHTS_NOT_SAMPLED_TYPE, NO_INSIGHTS_NOT_SAMPLED_SKIP_URI, NO_INSIGHTS_NOT_SAMPLED_URI_NOT_IN_LIST, NO_INSIGHTS_NOT_SAMPLED_CLIENT_IP_NOT_IN_RANGE, NO_INSIGHTS_NOT_SAMPLED_OTHER, ACTIVE_INSIGHTS_FAILED, ACTIVE_INSIGHTS_ENABLED, PASSIVE_INSIGHTS_ENABLED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientInsights *string `json:"client_insights,omitempty"`

	// IPv4 address of the client. When true client IP feature is enabled, this will be derived from the header configured in the true client IP feature, if present in the request. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientIP *uint32 `json:"client_ip"`

	// IPv6 address of the client. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIp6 *string `json:"client_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientLocation *string `json:"client_location,omitempty"`

	// Name of the Client Log Filter applied. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientLogFilterName *string `json:"client_log_filter_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientOs *string `json:"client_os,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientRtt *uint32 `json:"client_rtt"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientSrcPort *uint32 `json:"client_src_port"`

	//  Enum options - NO_COMPRESSION_DISABLED, NO_COMPRESSION_GZIP_CONTENT, NO_COMPRESSION_CONTENT_TYPE, NO_COMPRESSION_CUSTOM_FILTER, NO_COMPRESSION_AUTO_FILTER, NO_COMPRESSION_MIN_LENGTH, NO_COMPRESSION_CAN_BE_COMPRESSED, COMPRESSED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Compression *string `json:"compression,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CompressionPercentage *int32 `json:"compression_percentage,omitempty"`

	// TCP connection establishment time for the backend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ConnEstTimeBe uint32 `json:"conn_est_time_be,omitempty"`

	// TCP connection establishment time for the frontend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ConnEstTimeFe uint32 `json:"conn_est_time_fe,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnectionErrorInfo *ConnErrorInfo `json:"connection_error_info,omitempty"`

	// Critical error encountered during request processing. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CriticalErrorEncountered *bool `json:"critical_error_encountered,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DataTransferTime uint64 `json:"data_transfer_time,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DatascriptErrorTrace *DataScriptErrorTrace `json:"datascript_error_trace,omitempty"`

	// Log created by the invocations of the DataScript api avi.vs.log(). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DatascriptLog *string `json:"datascript_log,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Etag *string `json:"etag,omitempty"`

	// The method called by the gRPC request. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GrpcMethodName *string `json:"grpc_method_name,omitempty"`

	// The service called by the gRPC request. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GrpcServiceName *string `json:"grpc_service_name,omitempty"`

	// GRPC response status sent in the GRPC trailer. Special values are -1- No GRPC status recevied even though client sent content-type as application/grpc.. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GrpcStatus *int32 `json:"grpc_status,omitempty"`

	// The reason phrase corresponding to the gRPC status code. Enum options - GRPC_STATUS_CODE_OK, GRPC_STATUS_CODE_CANCELLED, GRPC_STATUS_CODE_UNKNOWN, GRPC_STATUS_CODE_INVALID_ARGUMENT, GRPC_STATUS_CODE_DEADLINE_EXCEEDED, GRPC_STATUS_CODE_NOT_FOUND, GRPC_STATUS_CODE_ALREADY_EXISTS, GRPC_STATUS_CODE_PERMISSION_DENIED, GRPC_STATUS_CODE_RESOURCE_EXHAUSTED, GRPC_STATUS_CODE_FAILED_PRECONDITION, GRPC_STATUS_CODE_STOPPED, GRPC_STATUS_CODE_OUT_OF_RANGE, GRPC_STATUS_CODE_UNIMPLEMENTED, GRPC_STATUS_CODE_INTERNAL, GRPC_STATUS_CODE_UNAVAILABLE, GRPC_STATUS_CODE_DATA_LOSS, GRPC_STATUS_CODE_UNAUTHENTICATED. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GrpcStatusReasonPhrase *string `json:"grpc_status_reason_phrase,omitempty"`

	// Response headers received from backend server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to backend server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Host *string `json:"host,omitempty"`

	// Stream identifier corresponding to an HTTP2 request. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Http2StreamID uint32 `json:"http2_stream_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPRequestPolicyRuleName *string `json:"http_request_policy_rule_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPResponsePolicyRuleName *string `json:"http_response_policy_rule_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPSecurityPolicyRuleName *string `json:"http_security_policy_rule_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPVersion *string `json:"http_version,omitempty"`

	// Log for the ICAP processing. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IcapLog *IcapLog `json:"icap_log,omitempty"`

	// Logs for the JWT Validation process. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	JwtLog *JwtLog `json:"jwt_log,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	LogID *uint32 `json:"log_id"`

	// Maximum packet processing latency for the backend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxIngressLatencyBe uint32 `json:"max_ingress_latency_be,omitempty"`

	// Maximum packet processing latency for the frontend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxIngressLatencyFe uint32 `json:"max_ingress_latency_fe,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Method *string `json:"method,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Microservice *string `json:"microservice,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MicroserviceName *string `json:"microservice_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkSecurityPolicyRuleName *string `json:"network_security_policy_rule_name,omitempty"`

	// NTLM auto-detection logs. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NtlmLog *NtlmLog `json:"ntlm_log,omitempty"`

	// Logs related to OAuth requests. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthLog *OauthLog `json:"oauth_log,omitempty"`

	// OCSP Certificate Status response sent in the SSL/TLS connection handshake. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OcspStatusRespSent *bool `json:"ocsp_status_resp_sent,omitempty"`

	// Logs for HTTP Out-Of-Band Requests. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OobLog *OutOfBandRequestLog `json:"oob_log,omitempty"`

	// The actual client request URI sent before normalization. Only included if it differs from the normalized URI. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OrigURI *string `json:"orig_uri,omitempty"`

	// Logs for the PingAccess authentication process. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PaaLog *PaaLog `json:"paa_log,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PersistenceUsed *bool `json:"persistence_used,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PersistentSessionID uint64 `json:"persistent_session_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Pool *string `json:"pool,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RedirectedURI *string `json:"redirected_uri,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Referer *string `json:"referer,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ReportTimestamp *uint64 `json:"report_timestamp"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestContentType *string `json:"request_content_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestHeaders uint32 `json:"request_headers,omitempty"`

	// Unique HTTP Request ID . Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestID *string `json:"request_id,omitempty"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestLength uint64 `json:"request_length,omitempty"`

	// Flag to indicate if request was served locally because the remote site was down. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestServedLocallyRemoteSiteDown *bool `json:"request_served_locally_remote_site_down,omitempty"`

	//  Enum options - AVI_HTTP_REQUEST_STATE_CONN_ACCEPT, AVI_HTTP_REQUEST_STATE_WAITING_FOR_REQUEST, AVI_HTTP_REQUEST_STATE_SSL_HANDSHAKING, AVI_HTTP_REQUEST_STATE_PROCESSING_SPDY, AVI_HTTP_REQUEST_STATE_READ_CLIENT_REQ_LINE, AVI_HTTP_REQUEST_STATE_READ_CLIENT_REQ_HDR, AVI_HTTP_REQUEST_STATE_CONNECT_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_SEND_REQ_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_READ_RESP_HDR_FROM_UPSTREAM, AVI_HTTP_REQUEST_STATE_SEND_TO_CLIENT, AVI_HTTP_REQUEST_STATE_KEEPALIVE, AVI_HTTP_REQUEST_STATE_PROXY_UPGRADED_CONN, AVI_HTTP_REQUEST_STATE_CLOSING_REQUEST, AVI_HTTP_REQUEST_STATE_READ_FROM_UPSTREAM, AVI_HTTP_REQUEST_STATE_READ_PROXY_PROTOCOL, AVI_HTTP_REQUEST_STATE_READ_CLIENT_PIPELINE_REQ_LINE, AVI_HTTP_REQUEST_STATE_SSL_HANDSHAKE_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_WAITING_IN_CONNPOOL_CACHE, AVI_HTTP_REQUEST_STATE_SEND_RESPONSE_HEADER_TO_CLIENT, AVI_HTTP_REQUEST_STATE_SEND_RESPONSE_BODY_TO_CLIENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestState *string `json:"request_state,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseCode uint32 `json:"response_code,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseContentType *string `json:"response_content_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseHeaders uint32 `json:"response_headers,omitempty"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseLength uint64 `json:"response_length,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseTimeFirstByte uint64 `json:"response_time_first_byte,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseTimeLastByte uint64 `json:"response_time_last_byte,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RewrittenURIPath *string `json:"rewritten_uri_path,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RewrittenURIQuery *string `json:"rewritten_uri_query,omitempty"`

	// SAML authentication request is generated. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthRequestGenerated *bool `json:"saml_auth_request_generated,omitempty"`

	// SAML authentication response is received. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthResponseReceived *bool `json:"saml_auth_response_received,omitempty"`

	// SAML authentication session ID. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthSessionID uint64 `json:"saml_auth_session_id,omitempty"`

	// SAML authentication is used. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlAuthenticationUsed *bool `json:"saml_authentication_used,omitempty"`

	// Logs for the SAML authentication/authorization process. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlLog *SamlLog `json:"saml_log,omitempty"`

	// SAML authentication session cookie is valid. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlSessionCookieValid *bool `json:"saml_session_cookie_valid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerConnSrcIP uint32 `json:"server_conn_src_ip,omitempty"`

	// IPv6 address used to connect to Server. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerConnSrcIp6 *string `json:"server_conn_src_ip6,omitempty"`

	// Flag to indicate if connection from the connection pool was reused. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerConnectionReused *bool `json:"server_connection_reused,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerDestPort uint32 `json:"server_dest_port,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerIP uint32 `json:"server_ip,omitempty"`

	// IPv6 address of the Server. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerIp6 *string `json:"server_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerName *string `json:"server_name,omitempty"`

	// Request which initiates Server Push. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerPushInitiated *bool `json:"server_push_initiated,omitempty"`

	// Requests served via Server Push. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerPushedRequest *bool `json:"server_pushed_request,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerResponseCode uint32 `json:"server_response_code,omitempty"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerResponseLength uint64 `json:"server_response_length,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerResponseTimeFirstByte uint64 `json:"server_response_time_first_byte,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerResponseTimeLastByte uint64 `json:"server_response_time_last_byte,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerRtt uint32 `json:"server_rtt,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerSideRedirectURI *string `json:"server_side_redirect_uri,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerSrcPort uint32 `json:"server_src_port,omitempty"`

	// SSL session id for the backend connection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerSslSessionID *string `json:"server_ssl_session_id,omitempty"`

	// Flag to indicate if SSL session was reused. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerSslSessionReused *bool `json:"server_ssl_session_reused,omitempty"`

	// Number of servers tried during server reselect before the response is sent back. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServersTried uint32 `json:"servers_tried,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServiceEngine *string `json:"service_engine"`

	// Field set by datascript using avi.vs.set_session_id(). Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionID *string `json:"session_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Significance *string `json:"significance,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Significant *uint64 `json:"significant"`

	// List of enums which indicate why a log is significant. Enum options - ADF_CLIENT_CONN_SETUP_REFUSED, ADF_SERVER_CONN_SETUP_REFUSED, ADF_CLIENT_CONN_SETUP_TIMEDOUT, ADF_SERVER_CONN_SETUP_TIMEDOUT, ADF_CLIENT_CONN_SETUP_FAILED_INTERNAL, ADF_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_CONN_SETUP_FAILED_BAD_PACKET, ADF_UDP_CONN_SETUP_FAILED_INTERNAL, ADF_UDP_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_SENT_RESET, ADF_SERVER_SENT_RESET, ADF_CLIENT_CONN_TIMEDOUT, ADF_SERVER_CONN_TIMEDOUT, ADF_USER_DELETE_OPERATION, ADF_CLIENT_REQUEST_TIMEOUT, ADF_CLIENT_CONN_ABORTED, ADF_CLIENT_SSL_HANDSHAKE_FAILURE, ADF_CLIENT_CONN_FAILED, ADF_SERVER_CERTIFICATE_VERIFICATION_FAILED, ADF_SERVER_SIDE_SSL_HANDSHAKE_FAILED.... Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SignificantLog []string `json:"significant_log,omitempty"`

	//  Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SniHostname *string `json:"sni_hostname,omitempty"`

	// Source IP of the client connection to the VS. This can be different from client IP when true client IP feature is enabled. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SourceIP uint32 `json:"source_ip,omitempty"`

	// IPv6 address of the source of the client connection to the VS. This can be different from client IPv6 address when true client IP feature is enabled. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SourceIp6 *string `json:"source_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SpdyVersion *string `json:"spdy_version,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslCipher *string `json:"ssl_cipher,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslSessionID *string `json:"ssl_session_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslVersion *string `json:"ssl_version,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TotalTime uint64 `json:"total_time,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Udf *bool `json:"udf"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIPath *string `json:"uri_path,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIQuery *string `json:"uri_query,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserAgent *string `json:"user_agent,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserID *string `json:"user_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcpuID *uint32 `json:"vcpu_id"`

	// EVH rule matching the request. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VhMatchRule *string `json:"vh_match_rule,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Virtualservice *string `json:"virtualservice"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsIP uint32 `json:"vs_ip,omitempty"`

	// Virtual IPv6 address of the VS. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsIp6 *string `json:"vs_ip6,omitempty"`

	// Presence of waf_log indicates that atleast 1 WAF rule was hit for the transaction. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WafLog *WafLog `json:"waf_log,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Xff *string `json:"xff,omitempty"`
}

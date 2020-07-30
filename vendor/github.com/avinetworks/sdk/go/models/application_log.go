package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApplicationLog application log
// swagger:model ApplicationLog
type ApplicationLog struct {

	// Placeholder for description of property adf of obj type ApplicationLog field type str  type boolean
	// Required: true
	Adf *bool `json:"adf"`

	// all_request_headers of ApplicationLog.
	AllRequestHeaders *string `json:"all_request_headers,omitempty"`

	// all_response_headers of ApplicationLog.
	AllResponseHeaders *string `json:"all_response_headers,omitempty"`

	// Number of app_response_time.
	AppResponseTime *int64 `json:"app_response_time,omitempty"`

	//  Enum options - NOT_UPDATED, BY_CONTENT_REWRITE_PROFILE, BY_DATA_SCRIPT. Field introduced in 17.1.1.
	BodyUpdated *string `json:"body_updated,omitempty"`

	// Cache fetch and store is disabled by the Datascript policies. Field introduced in 20.1.1.
	CacheDisabledByDs *bool `json:"cache_disabled_by_ds,omitempty"`

	// Placeholder for description of property cache_hit of obj type ApplicationLog field type str  type boolean
	CacheHit *bool `json:"cache_hit,omitempty"`

	// Placeholder for description of property cacheable of obj type ApplicationLog field type str  type boolean
	Cacheable *bool `json:"cacheable,omitempty"`

	// Byte stream of client cipher list sent on SSL_R_NO_SHARED_CIPHER error. Field introduced in 18.1.4, 18.2.1.
	CipherBytes *string `json:"cipher_bytes,omitempty"`

	// client_browser of ApplicationLog.
	ClientBrowser *string `json:"client_browser,omitempty"`

	// List of ciphers sent by client in TLS/SSL Client Hello. Only sent when TLS handshake fails due to no shared cipher. Field introduced in 18.1.4, 18.2.1.
	ClientCipherList *SSLCipherList `json:"client_cipher_list,omitempty"`

	// Number of client_dest_port.
	// Required: true
	ClientDestPort *int32 `json:"client_dest_port"`

	// client_device of ApplicationLog.
	ClientDevice *string `json:"client_device,omitempty"`

	//  Enum options - INSIGHTS_DISABLED, NO_INSIGHTS_NOT_SAMPLED_COUNT, NO_INSIGHTS_NOT_SAMPLED_TYPE, NO_INSIGHTS_NOT_SAMPLED_SKIP_URI, NO_INSIGHTS_NOT_SAMPLED_URI_NOT_IN_LIST, NO_INSIGHTS_NOT_SAMPLED_CLIENT_IP_NOT_IN_RANGE, NO_INSIGHTS_NOT_SAMPLED_OTHER, ACTIVE_INSIGHTS_FAILED, ACTIVE_INSIGHTS_ENABLED, PASSIVE_INSIGHTS_ENABLED.
	ClientInsights *string `json:"client_insights,omitempty"`

	// Number of client_ip.
	// Required: true
	ClientIP *int32 `json:"client_ip"`

	// IPv6 address of the client. Field introduced in 18.1.1.
	ClientIp6 *string `json:"client_ip6,omitempty"`

	// client_location of ApplicationLog.
	ClientLocation *string `json:"client_location,omitempty"`

	// Name of the Client Log Filter applied. Field introduced in 18.1.5, 18.2.1.
	ClientLogFilterName *string `json:"client_log_filter_name,omitempty"`

	// client_os of ApplicationLog.
	ClientOs *string `json:"client_os,omitempty"`

	// Number of client_rtt.
	// Required: true
	ClientRtt *int32 `json:"client_rtt"`

	// Number of client_src_port.
	// Required: true
	ClientSrcPort *int32 `json:"client_src_port"`

	//  Enum options - NO_COMPRESSION_DISABLED, NO_COMPRESSION_GZIP_CONTENT, NO_COMPRESSION_CONTENT_TYPE, NO_COMPRESSION_CUSTOM_FILTER, NO_COMPRESSION_AUTO_FILTER, NO_COMPRESSION_MIN_LENGTH, NO_COMPRESSION_CAN_BE_COMPRESSED, COMPRESSED.
	Compression *string `json:"compression,omitempty"`

	// Number of compression_percentage.
	CompressionPercentage *int32 `json:"compression_percentage,omitempty"`

	// Placeholder for description of property connection_error_info of obj type ApplicationLog field type str  type object
	ConnectionErrorInfo *ConnErrorInfo `json:"connection_error_info,omitempty"`

	// Number of data_transfer_time.
	DataTransferTime *int64 `json:"data_transfer_time,omitempty"`

	// Placeholder for description of property datascript_error_trace of obj type ApplicationLog field type str  type object
	DatascriptErrorTrace *DataScriptErrorTrace `json:"datascript_error_trace,omitempty"`

	// Log created by the invocations of the DataScript api avi.vs.log().
	DatascriptLog *string `json:"datascript_log,omitempty"`

	// etag of ApplicationLog.
	Etag *string `json:"etag,omitempty"`

	// The method called by the gRPC request. Field introduced in 20.1.1.
	GrpcMethodName *string `json:"grpc_method_name,omitempty"`

	// The service called by the gRPC request. Field introduced in 20.1.1.
	GrpcServiceName *string `json:"grpc_service_name,omitempty"`

	// GRPC response status sent in the GRPC trailer. Special values are -1- 'No GRPC status recevied even though client sent content-type as application/grpc.'. Field introduced in 20.1.1.
	GrpcStatus *int32 `json:"grpc_status,omitempty"`

	// The reason phrase corresponding to the gRPC status code. Enum options - GRPC_STATUS_CODE_OK, GRPC_STATUS_CODE_CANCELLED, GRPC_STATUS_CODE_UNKNOWN, GRPC_STATUS_CODE_INVALID_ARGUMENT, GRPC_STATUS_CODE_DEADLINE_EXCEEDED, GRPC_STATUS_CODE_NOT_FOUND, GRPC_STATUS_CODE_ALREADY_EXISTS, GRPC_STATUS_CODE_PERMISSION_DENIED, GRPC_STATUS_CODE_UNAUTHENTICATED, GRPC_STATUS_CODE_RESOURCE_EXHAUSTED, GRPC_STATUS_CODE_FAILED_PRECONDITION, GRPC_STATUS_CODE_ABORTED, GRPC_STATUS_CODE_OUT_OF_RANGE, GRPC_STATUS_CODE_UNIMPLEMENTED, GRPC_STATUS_CODE_INTERNAL, GRPC_STATUS_CODE_UNAVAILABLE, GRPC_STATUS_CODE_DATA_LOSS. Field introduced in 20.1.1.
	GrpcStatusReasonPhrase *string `json:"grpc_status_reason_phrase,omitempty"`

	// Response headers received from backend server.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to backend server.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	// host of ApplicationLog.
	Host *string `json:"host,omitempty"`

	// Stream identifier corresponding to an HTTP2 request. Field introduced in 18.1.2.
	Http2StreamID *int32 `json:"http2_stream_id,omitempty"`

	// http_request_policy_rule_name of ApplicationLog.
	HTTPRequestPolicyRuleName *string `json:"http_request_policy_rule_name,omitempty"`

	// http_response_policy_rule_name of ApplicationLog.
	HTTPResponsePolicyRuleName *string `json:"http_response_policy_rule_name,omitempty"`

	// http_security_policy_rule_name of ApplicationLog.
	HTTPSecurityPolicyRuleName *string `json:"http_security_policy_rule_name,omitempty"`

	// http_version of ApplicationLog.
	HTTPVersion *string `json:"http_version,omitempty"`

	// Log for the ICAP processing. Field introduced in 20.1.1.
	IcapLog *IcapLog `json:"icap_log,omitempty"`

	// Number of log_id.
	// Required: true
	LogID *int32 `json:"log_id"`

	// method of ApplicationLog.
	Method *string `json:"method,omitempty"`

	// microservice of ApplicationLog.
	Microservice *string `json:"microservice,omitempty"`

	// microservice_name of ApplicationLog.
	MicroserviceName *string `json:"microservice_name,omitempty"`

	// network_security_policy_rule_name of ApplicationLog.
	NetworkSecurityPolicyRuleName *string `json:"network_security_policy_rule_name,omitempty"`

	// OCSP Certificate Status response sent in the SSL/TLS connection handshake. Field introduced in 20.1.1.
	OcspStatusRespSent *bool `json:"ocsp_status_resp_sent,omitempty"`

	// Logs for the PingAccess authentication process. Field introduced in 18.2.3.
	PaaLog *PaaLog `json:"paa_log,omitempty"`

	// Placeholder for description of property persistence_used of obj type ApplicationLog field type str  type boolean
	PersistenceUsed *bool `json:"persistence_used,omitempty"`

	// Number of persistent_session_id.
	PersistentSessionID *int64 `json:"persistent_session_id,omitempty"`

	// pool of ApplicationLog.
	Pool *string `json:"pool,omitempty"`

	// pool_name of ApplicationLog.
	PoolName *string `json:"pool_name,omitempty"`

	// redirected_uri of ApplicationLog.
	RedirectedURI *string `json:"redirected_uri,omitempty"`

	// referer of ApplicationLog.
	Referer *string `json:"referer,omitempty"`

	// Number of report_timestamp.
	// Required: true
	ReportTimestamp *int64 `json:"report_timestamp"`

	// request_content_type of ApplicationLog.
	RequestContentType *string `json:"request_content_type,omitempty"`

	// Number of request_headers.
	RequestHeaders *int32 `json:"request_headers,omitempty"`

	// Unique HTTP Request ID . Field introduced in 17.2.4.
	RequestID *string `json:"request_id,omitempty"`

	// Number of request_length.
	RequestLength *int64 `json:"request_length,omitempty"`

	// Flag to indicate if request was served locally because the remote site was down. Field introduced in 17.2.5.
	RequestServedLocallyRemoteSiteDown *bool `json:"request_served_locally_remote_site_down,omitempty"`

	//  Enum options - AVI_HTTP_REQUEST_STATE_CONN_ACCEPT, AVI_HTTP_REQUEST_STATE_WAITING_FOR_REQUEST, AVI_HTTP_REQUEST_STATE_SSL_HANDSHAKING, AVI_HTTP_REQUEST_STATE_PROCESSING_SPDY, AVI_HTTP_REQUEST_STATE_READ_CLIENT_REQ_LINE, AVI_HTTP_REQUEST_STATE_READ_CLIENT_REQ_HDR, AVI_HTTP_REQUEST_STATE_CONNECT_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_SEND_REQ_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_READ_RESP_HDR_FROM_UPSTREAM, AVI_HTTP_REQUEST_STATE_SEND_TO_CLIENT, AVI_HTTP_REQUEST_STATE_KEEPALIVE, AVI_HTTP_REQUEST_STATE_PROXY_UPGRADED_CONN, AVI_HTTP_REQUEST_STATE_CLOSING_REQUEST, AVI_HTTP_REQUEST_STATE_READ_FROM_UPSTREAM, AVI_HTTP_REQUEST_STATE_READ_PROXY_PROTOCOL, AVI_HTTP_REQUEST_STATE_READ_CLIENT_PIPELINE_REQ_LINE, AVI_HTTP_REQUEST_STATE_SSL_HANDSHAKE_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_WAITING_IN_CONNPOOL_CACHE, AVI_HTTP_REQUEST_STATE_SEND_RESPONSE_HEADER_TO_CLIENT, AVI_HTTP_REQUEST_STATE_SEND_RESPONSE_BODY_TO_CLIENT.
	RequestState *string `json:"request_state,omitempty"`

	// Number of response_code.
	ResponseCode *int32 `json:"response_code,omitempty"`

	// response_content_type of ApplicationLog.
	ResponseContentType *string `json:"response_content_type,omitempty"`

	// Number of response_headers.
	ResponseHeaders *int32 `json:"response_headers,omitempty"`

	// Number of response_length.
	ResponseLength *int64 `json:"response_length,omitempty"`

	// Number of response_time_first_byte.
	ResponseTimeFirstByte *int64 `json:"response_time_first_byte,omitempty"`

	// Number of response_time_last_byte.
	ResponseTimeLastByte *int64 `json:"response_time_last_byte,omitempty"`

	// rewritten_uri_path of ApplicationLog.
	RewrittenURIPath *string `json:"rewritten_uri_path,omitempty"`

	// rewritten_uri_query of ApplicationLog.
	RewrittenURIQuery *string `json:"rewritten_uri_query,omitempty"`

	// SAML authentication request is generated. Field introduced in 18.2.1.
	SamlAuthRequestGenerated *bool `json:"saml_auth_request_generated,omitempty"`

	// SAML authentication response is received. Field introduced in 18.2.1.
	SamlAuthResponseReceived *bool `json:"saml_auth_response_received,omitempty"`

	// SAML authentication session ID. Field introduced in 18.2.1.
	SamlAuthSessionID *int64 `json:"saml_auth_session_id,omitempty"`

	// SAML authentication is used. Field introduced in 18.2.1.
	SamlAuthenticationUsed *bool `json:"saml_authentication_used,omitempty"`

	// Logs for the SAML authentication/authorization process. Field introduced in 20.1.1.
	SamlLog *SamlLog `json:"saml_log,omitempty"`

	// SAML authentication session cookie is valid. Field introduced in 18.2.1.
	SamlSessionCookieValid *bool `json:"saml_session_cookie_valid,omitempty"`

	// Number of server_conn_src_ip.
	ServerConnSrcIP *int32 `json:"server_conn_src_ip,omitempty"`

	// IPv6 address used to connect to Server. Field introduced in 18.1.1.
	ServerConnSrcIp6 *string `json:"server_conn_src_ip6,omitempty"`

	// Flag to indicate if connection from the connection pool was reused.
	ServerConnectionReused *bool `json:"server_connection_reused,omitempty"`

	// Number of server_dest_port.
	ServerDestPort *int32 `json:"server_dest_port,omitempty"`

	// Number of server_ip.
	ServerIP *int32 `json:"server_ip,omitempty"`

	// IPv6 address of the Server. Field introduced in 18.1.1.
	ServerIp6 *string `json:"server_ip6,omitempty"`

	// server_name of ApplicationLog.
	ServerName *string `json:"server_name,omitempty"`

	// Number of server_response_code.
	ServerResponseCode *int32 `json:"server_response_code,omitempty"`

	// Number of server_response_length.
	ServerResponseLength *int64 `json:"server_response_length,omitempty"`

	// Number of server_response_time_first_byte.
	ServerResponseTimeFirstByte *int64 `json:"server_response_time_first_byte,omitempty"`

	// Number of server_response_time_last_byte.
	ServerResponseTimeLastByte *int64 `json:"server_response_time_last_byte,omitempty"`

	// Number of server_rtt.
	ServerRtt *int32 `json:"server_rtt,omitempty"`

	// server_side_redirect_uri of ApplicationLog.
	ServerSideRedirectURI *string `json:"server_side_redirect_uri,omitempty"`

	// Number of server_src_port.
	ServerSrcPort *int32 `json:"server_src_port,omitempty"`

	// SSL session id for the backend connection.
	ServerSslSessionID *string `json:"server_ssl_session_id,omitempty"`

	// Flag to indicate if SSL session was reused.
	ServerSslSessionReused *bool `json:"server_ssl_session_reused,omitempty"`

	// Number of servers tried during server reselect before the response is sent back. Field introduced in 18.2.2.
	ServersTried *int32 `json:"servers_tried,omitempty"`

	// service_engine of ApplicationLog.
	// Required: true
	ServiceEngine *string `json:"service_engine"`

	// significance of ApplicationLog.
	Significance *string `json:"significance,omitempty"`

	// Number of significant.
	// Required: true
	Significant *int64 `json:"significant"`

	// List of enums which indicate why a log is significant. Enum options - ADF_CLIENT_CONN_SETUP_REFUSED, ADF_SERVER_CONN_SETUP_REFUSED, ADF_CLIENT_CONN_SETUP_TIMEDOUT, ADF_SERVER_CONN_SETUP_TIMEDOUT, ADF_CLIENT_CONN_SETUP_FAILED_INTERNAL, ADF_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_CONN_SETUP_FAILED_BAD_PACKET, ADF_UDP_CONN_SETUP_FAILED_INTERNAL, ADF_UDP_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_SENT_RESET, ADF_SERVER_SENT_RESET, ADF_CLIENT_CONN_TIMEDOUT, ADF_SERVER_CONN_TIMEDOUT, ADF_USER_DELETE_OPERATION, ADF_CLIENT_REQUEST_TIMEOUT, ADF_CLIENT_CONN_ABORTED, ADF_CLIENT_SSL_HANDSHAKE_FAILURE, ADF_CLIENT_CONN_FAILED, ADF_SERVER_CERTIFICATE_VERIFICATION_FAILED, ADF_SERVER_SIDE_SSL_HANDSHAKE_FAILED...
	SignificantLog []string `json:"significant_log,omitempty"`

	//  Field introduced in 17.2.5.
	SniHostname *string `json:"sni_hostname,omitempty"`

	// spdy_version of ApplicationLog.
	SpdyVersion *string `json:"spdy_version,omitempty"`

	// ssl_cipher of ApplicationLog.
	SslCipher *string `json:"ssl_cipher,omitempty"`

	// ssl_session_id of ApplicationLog.
	SslSessionID *string `json:"ssl_session_id,omitempty"`

	// ssl_version of ApplicationLog.
	SslVersion *string `json:"ssl_version,omitempty"`

	// Number of total_time.
	TotalTime *int64 `json:"total_time,omitempty"`

	// Placeholder for description of property udf of obj type ApplicationLog field type str  type boolean
	// Required: true
	Udf *bool `json:"udf"`

	// uri_path of ApplicationLog.
	URIPath *string `json:"uri_path,omitempty"`

	// uri_query of ApplicationLog.
	URIQuery *string `json:"uri_query,omitempty"`

	// user_agent of ApplicationLog.
	UserAgent *string `json:"user_agent,omitempty"`

	// user_id of ApplicationLog.
	UserID *string `json:"user_id,omitempty"`

	// Number of vcpu_id.
	// Required: true
	VcpuID *int32 `json:"vcpu_id"`

	// virtualservice of ApplicationLog.
	// Required: true
	Virtualservice *string `json:"virtualservice"`

	//  Field introduced in 17.1.1.
	VsIP *int32 `json:"vs_ip,omitempty"`

	// Virtual IPv6 address of the VS. Field introduced in 18.1.1.
	VsIp6 *string `json:"vs_ip6,omitempty"`

	// Presence of waf_log indicates that atleast 1 WAF rule was hit for the transaction. Field introduced in 17.2.1.
	WafLog *WafLog `json:"waf_log,omitempty"`

	// xff of ApplicationLog.
	Xff *string `json:"xff,omitempty"`
}

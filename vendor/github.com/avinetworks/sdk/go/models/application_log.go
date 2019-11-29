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

	// Response headers received from backend server.
	HeadersReceivedFromServer *string `json:"headers_received_from_server,omitempty"`

	// Request headers sent to backend server.
	HeadersSentToServer *string `json:"headers_sent_to_server,omitempty"`

	// host of ApplicationLog.
	Host *string `json:"host,omitempty"`

	// Stream identifier corresponding to an HTTP/2 request. Field introduced in 18.1.2.
	Http2StreamID *int32 `json:"http2_stream_id,omitempty"`

	// http_request_policy_rule_name of ApplicationLog.
	HTTPRequestPolicyRuleName *string `json:"http_request_policy_rule_name,omitempty"`

	// http_response_policy_rule_name of ApplicationLog.
	HTTPResponsePolicyRuleName *string `json:"http_response_policy_rule_name,omitempty"`

	// http_security_policy_rule_name of ApplicationLog.
	HTTPSecurityPolicyRuleName *string `json:"http_security_policy_rule_name,omitempty"`

	// http_version of ApplicationLog.
	HTTPVersion *string `json:"http_version,omitempty"`

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

	//  Enum options - AVI_HTTP_REQUEST_STATE_CONN_ACCEPT, AVI_HTTP_REQUEST_STATE_WAITING_FOR_REQUEST, AVI_HTTP_REQUEST_STATE_SSL_HANDSHAKING, AVI_HTTP_REQUEST_STATE_PROCESSING_SPDY, AVI_HTTP_REQUEST_STATE_READ_CLIENT_REQ_LINE, AVI_HTTP_REQUEST_STATE_READ_CLIENT_REQ_HDR, AVI_HTTP_REQUEST_STATE_CONNECT_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_SEND_REQ_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_READ_RESP_HDR_FROM_UPSTREAM, AVI_HTTP_REQUEST_STATE_SEND_TO_CLIENT, AVI_HTTP_REQUEST_STATE_KEEPALIVE, AVI_HTTP_REQUEST_STATE_PROXY_UPGRADED_CONN, AVI_HTTP_REQUEST_STATE_CLOSING_REQUEST, AVI_HTTP_REQUEST_STATE_READ_FROM_UPSTREAM, AVI_HTTP_REQUEST_STATE_READ_PROXY_PROTOCOL, AVI_HTTP_REQUEST_STATE_READ_CLIENT_PIPELINE_REQ_LINE, AVI_HTTP_REQUEST_STATE_SSL_HANDSHAKE_TO_UPSTREAM, AVI_HTTP_REQUEST_STATE_WAITING_IN_CONNPOOL_CACHE.
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

	// service_engine of ApplicationLog.
	// Required: true
	ServiceEngine *string `json:"service_engine"`

	// significance of ApplicationLog.
	Significance *string `json:"significance,omitempty"`

	// Number of significant.
	// Required: true
	Significant *int64 `json:"significant"`

	// List of enums which indicate why a log is significant. Enum options - ADF_CLIENT_CONN_SETUP_REFUSED, ADF_SERVER_CONN_SETUP_REFUSED, ADF_CLIENT_CONN_SETUP_TIMEDOUT, ADF_SERVER_CONN_SETUP_TIMEDOUT, ADF_CLIENT_CONN_SETUP_FAILED_INTERNAL, ADF_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_CONN_SETUP_FAILED_BAD_PACKET, ADF_UDP_CONN_SETUP_FAILED_INTERNAL, ADF_UDP_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_SENT_RESET, ADF_SERVER_SENT_RESET, ADF_CLIENT_CONN_TIMEDOUT, ADF_SERVER_CONN_TIMEDOUT, ADF_USER_DELETE_OPERATION, ADF_CLIENT_REQUEST_TIMEOUT, ADF_CLIENT_CONN_ABORTED, ADF_CLIENT_SSL_HANDSHAKE_FAILURE, ADF_CLIENT_CONN_FAILED, ADF_SERVER_CERTIFICATE_VERIFICATION_FAILED, ADF_SERVER_SIDE_SSL_HANDSHAKE_FAILED, ADF_IDLE_TIMEDOUT, ADF_CLIENT_CONNECTION_CLOSED_BEFORE_REQUEST, ADF_CLIENT_HIGH_TIMEOUT_RETRANSMITS, ADF_SERVER_HIGH_TIMEOUT_RETRANSMITS, ADF_CLIENT_HIGH_RX_ZERO_WINDOW_SIZE_EVENTS, ADF_SERVER_HIGH_RX_ZERO_WINDOW_SIZE_EVENTS, ADF_CLIENT_RTT_ABOVE_SEC, ADF_SERVER_RTT_ABOVE_500MS, ADF_CLIENT_HIGH_TOTAL_RETRANSMITS, ADF_SERVER_HIGH_TOTAL_RETRANSMITS, ADF_CLIENT_HIGH_OUT_OF_ORDERS, ADF_SERVER_HIGH_OUT_OF_ORDERS, ADF_CLIENT_HIGH_TX_ZERO_WINDOW_SIZE_EVENTS, ADF_SERVER_HIGH_TX_ZERO_WINDOW_SIZE_EVENTS, ADF_CLIENT_POSSIBLE_WINDOW_STUCK, ADF_SERVER_POSSIBLE_WINDOW_STUCK, ADF_SERVER_UNANSWERED_SYNS, ADF_CLIENT_CLOSE_CONNECTION_ON_VS_UPDATE, ADF_RESPONSE_CODE_4XX, ADF_RESPONSE_CODE_5XX, ADF_LOAD_BALANCING_FAILED, ADF_DATASCRIPT_EXECUTION_FAILED, ADF_REQUEST_NO_POOL, ADF_RATE_LIMIT_DROP_CLIENT_IP, ADF_RATE_LIMIT_DROP_URI, ADF_RATE_LIMIT_DROP_CLIENT_IP_URI, ADF_RATE_LIMIT_DROP_UNKNOWN_URI, ADF_RATE_LIMIT_DROP_BAD_URI, ADF_REQUEST_VIRTUAL_HOSTING_APP_SELECT_FAILED, ADF_RATE_LIMIT_DROP_UNKNOWN_CIP, ADF_RATE_LIMIT_DROP_BAD_CIP, ADF_RATE_LIMIT_DROP_CLIENT_IP_BAD, ADF_RATE_LIMIT_DROP_URI_BAD, ADF_RATE_LIMIT_DROP_CLIENT_IP_URI_BAD, ADF_RATE_LIMIT_DROP_REQ, ADF_RATE_LIMIT_DROP_CLIENT_IP_CONN, ADF_RATE_LIMIT_DROP_CONN, ADF_RATE_LIMIT_DROP_HEADER, ADF_RATE_LIMIT_DROP_CUSTOM, ADF_HTTP_VERSION_LT_1_0, ADF_CLIENT_HIGH_RESPONSE_TIME, ADF_SERVER_HIGH_RESPONSE_TIME, ADF_PERSISTENT_SERVER_CHANGE, ADF_DOS_SERVER_BAD_GATEWAY, ADF_DOS_SERVER_GATEWAY_TIMEOUT, ADF_DOS_CLIENT_SENT_RESET, ADF_DOS_CLIENT_CONN_TIMEOUT, ADF_DOS_CLIENT_REQUEST_TIMEOUT, ADF_DOS_CLIENT_CONN_ABORTED, ADF_DOS_CLIENT_BAD_REQUEST, ADF_DOS_CLIENT_REQUEST_ENTITY_TOO_LARGE, ADF_DOS_CLIENT_REQUEST_URI_TOO_LARGE, ADF_DOS_CLIENT_REQUEST_HEADER_TOO_LARGE, ADF_DOS_CLIENT_CLOSED_REQUEST, ADF_DOS_SSL_ERROR, ADF_REQUEST_MEMORY_LIMIT_EXCEEDED, ADF_X509_CLIENT_CERTIFICATE_VERIFICATION_FAILED, ADF_X509_CLIENT_CERTIFICATE_NOT_YET_VALID, ADF_X509_CLIENT_CERTIFICATE_EXPIRED, ADF_X509_CLIENT_CERTIFICATE_REVOKED, ADF_X509_CLIENT_CERTIFICATE_INVALID_CA, ADF_X509_CLIENT_CERTIFICATE_CRL_NOT_PRESENT, ADF_X509_CLIENT_CERTIFICATE_CRL_NOT_YET_VALID, ADF_X509_CLIENT_CERTIFICATE_CRL_EXPIRED, ADF_X509_CLIENT_CERTIFICATE_CRL_ERROR, ADF_X509_CLIENT_CERTIFICATE_CHAINING_ERROR, ADF_X509_CLIENT_CERTIFICATE_INTERNAL_ERROR, ADF_X509_CLIENT_CERTIFICATE_FORMAT_ERROR, ADF_UDP_PORT_NOT_REACHABLE, ADF_UDP_CONN_TIMEOUT, ADF_X509_SERVER_CERTIFICATE_VERIFICATION_FAILED, ADF_X509_SERVER_CERTIFICATE_NOT_YET_VALID, ADF_X509_SERVER_CERTIFICATE_EXPIRED, ADF_X509_SERVER_CERTIFICATE_REVOKED, ADF_X509_SERVER_CERTIFICATE_INVALID_CA, ADF_X509_SERVER_CERTIFICATE_CRL_NOT_PRESENT, ADF_X509_SERVER_CERTIFICATE_CRL_NOT_YET_VALID, ADF_X509_SERVER_CERTIFICATE_CRL_EXPIRED, ADF_X509_SERVER_CERTIFICATE_CRL_ERROR, ADF_X509_SERVER_CERTIFICATE_CHAINING_ERROR, ADF_X509_SERVER_CERTIFICATE_INTERNAL_ERROR, ADF_X509_SERVER_CERTIFICATE_FORMAT_ERROR, ADF_X509_SERVER_CERTIFICATE_HOSTNAME_ERROR, ADF_SSL_R_BAD_CHANGE_CIPHER_SPEC, ADF_SSL_R_BLOCK_CIPHER_PAD_IS_WRONG, ADF_SSL_R_DIGEST_CHECK_FAILED, ADF_SSL_R_ERROR_IN_RECEIVED_CIPHER_LIST, ADF_SSL_R_EXCESSIVE_MESSAGE_SIZE, ADF_SSL_R_LENGTH_MISMATCH, ADF_SSL_R_NO_CIPHERS_PASSED, ADF_SSL_R_NO_CIPHERS_SPECIFIED, ADF_SSL_R_NO_COMPRESSION_SPECIFIED, ADF_SSL_R_NO_SHARED_CIPHER, ADF_SSL_R_RECORD_LENGTH_MISMATCH, ADF_SSL_R_PARSE_TLSEXT, ADF_SSL_R_UNEXPECTED_MESSAGE, ADF_SSL_R_UNEXPECTED_RECORD, ADF_SSL_R_UNKNOWN_ALERT_TYPE, ADF_SSL_R_UNKNOWN_PROTOCOL, ADF_SSL_R_WRONG_VERSION_NUMBER, ADF_SSL_R_DECRYPTION_FAILED_OR_BAD_RECORD_MAC, ADF_SSL_R_RENEGOTIATE_EXT_TOO_LONG, ADF_SSL_R_RENEGOTIATION_ENCODING_ERR, ADF_SSL_R_RENEGOTIATION_MISMATCH, ADF_SSL_R_UNSAFE_LEGACY_RENEGOTIATION_DISABLED, ADF_SSL_R_SCSV_RECEIVED_WHEN_RENEGOTIATING, ADF_SSL_R_INAPPROPRIATE_FALLBACK, ADF_SSL_R_SSLV3_ALERT_UNEXPECTED_MESSAGE, ADF_SSL_R_SSLV3_ALERT_BAD_RECORD_MAC, ADF_SSL_R_TLSV1_ALERT_DECRYPTION_FAILED, ADF_SSL_R_TLSV1_ALERT_RECORD_OVERFLOW, ADF_SSL_R_SSLV3_ALERT_DECOMPRESSION_FAILURE, ADF_SSL_R_SSLV3_ALERT_HANDSHAKE_FAILURE, ADF_SSL_R_SSLV3_ALERT_NO_CERTIFICATE, ADF_SSL_R_SSLV3_ALERT_BAD_CERTIFICATE, ADF_SSL_R_SSLV3_ALERT_UNSUPPORTED_CERTIFICATE, ADF_SSL_R_SSLV3_ALERT_CERTIFICATE_REVOKED, ADF_SSL_R_SSLV3_ALERT_CERTIFICATE_EXPIRED, ADF_SSL_R_SSLV3_ALERT_CERTIFICATE_UNKNOWN, ADF_SSL_R_SSLV3_ALERT_ILLEGAL_PARAMETER, ADF_SSL_R_TLSV1_ALERT_UNKNOWN_CA, ADF_SSL_R_TLSV1_ALERT_ACCESS_DENIED, ADF_SSL_R_TLSV1_ALERT_DECODE_ERROR, ADF_SSL_R_TLSV1_ALERT_DECRYPT_ERROR, ADF_SSL_R_TLSV1_ALERT_EXPORT_RESTRICTION, ADF_SSL_R_TLSV1_ALERT_PROTOCOL_VERSION, ADF_SSL_R_TLSV1_ALERT_INSUFFICIENT_SECURITY, ADF_SSL_R_TLSV1_ALERT_INTERNAL_ERROR, ADF_SSL_R_TLSV1_ALERT_USER_CANCELLED, ADF_SSL_R_TLSV1_ALERT_NO_RENEGOTIATION, ADF_CLIENT_AUTH_UNKNOWN_USER, ADF_CLIENT_AUTH_LOGIN_FAILED, ADF_CLIENT_AUTH_MISSING_CREDENTIALS, ADF_CLIENT_AUTH_SERVER_CONN_ERROR, ADF_CLIENT_AUTH_USER_NOT_AUTHORIZED, ADF_CLIENT_AUTH_TIMED_OUT, ADF_CLIENT_AUTH_UNKNOWN_ERROR, ADF_CLIENT_DNS_FAILED_INVALID_QUERY, ADF_CLIENT_DNS_FAILED_INVALID_DOMAIN, ADF_CLIENT_DNS_FAILED_NO_SERVICE, ADF_CLIENT_DNS_FAILED_GS_DOWN, ADF_CLIENT_DNS_FAILED_NO_VALID_GS_MEMBER, ADF_SERVER_DNS_ERROR_RESPONSE, ADF_CLIENT_DNS_FAILED_UNSUPPORTED_QUERY, ADF_MEMORY_EXHAUSTED, ADF_CLIENT_DNS_POLICY_DROP, ADF_WAF_MATCH, ADF_HTTP2_CLIENT_TIMEDOUT, ADF_HTTP2_PROXY_PROTOCOL_ERROR, ADF_HTTP2_INVALID_CONNECTION_PREFACE, ADF_HTTP2_CLIENT_INVALID_DATA_FRAME_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_PADDED_DATA_FRAME_WITH_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_VIOLATED_CONN_FLOW_CONTROL, ADF_HTTP2_CLIENT_VIOLATED_STREAM_FLOW_CONTROL, ADF_HTTP2_CLIENT_DATA_FRAME_HALF_CLOSED_STREAM, ADF_HTTP2_CLIENT_HEADERS_FRAME_WITH_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_HEADERS_FRAME_WITH_EMPTY_HEADER_BLOCK, ADF_HTTP2_CLIENT_PADDED_HEADERS_FRAME_WITH_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_HEADERS_FRAME_INCORRECT_IDENTIFIER, ADF_HTTP2_CLIENT_HEADERS_FRAME_STREAM_INCORRECT_DEPENDENCY, ADF_HTTP2_CONCURRENT_STREAMS_EXCEEDED, ADF_HTTP2_CLIENT_STREAM_DATA_BEFORE_ACK_SETTINGS, ADF_HTTP2_CLIENT_HEADER_BLOCK_TOO_LONG_SIZE_UPDATE, ADF_HTTP2_CLIENT_HEADER_BLOCK_TOO_LONG_HEADER_INDEX, ADF_HTTP2_CLIENT_HEADER_BLOCK_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_INVALID_HPACK_TABLE_INDEX, ADF_HTTP2_CLIENT_OUT_OF_BOUND_HPACK_TABLE_INDEX, ADF_HTTP2_CLIENT_INVALID_TABLE_SIZE_UPDATE, ADF_HTTP2_CLIENT_HEADER_FIELD_TOO_LONG_LENGTH_VALUE, ADF_HTTP2_CLIENT_EXCEEDED_HTTP2_MAX_FIELD_SIZE_LIMIT, ADF_HTTP2_CLIENT_INVALID_ENCODED_HEADER_FIELD, ADF_HTTP2_CLIENT_EXCEEDED_HTTP2_MAX_HEADER_SIZE_LIMIT, ADF_HTTP2_CLIENT_INVALID_HEADER_NAME, ADF_HTTP2_CLIENT_HEADER_WITH_INVALID_VALUE, ADF_HTTP2_CLIENT_UNKNOWN_PSEUDO_HEADER, ADF_HTTP2_CLIENT_DUPLICATE_PATH_HEADER, ADF_HTTP2_CLIENT_EMPTY_PATH_HEADER, ADF_HTTP2_CLIENT_INVALID_PATH_HEADER, ADF_HTTP2_CLIENT_DUPLICATE_METHOD_HEADER, ADF_HTTP2_CLIENT_EMPTY_METHOD_HEADER, ADF_HTTP2_CLIENT_INVALID_METHOD_HEADER, ADF_HTTP2_CLIENT_DUPLICATE_SCHEME_HEADER, ADF_HTTP2_CLIENT_EMPTY_SCHEME_HEADER, ADF_HTTP2_CLIENT_NO_METHOD_HEADER, ADF_HTTP2_CLIENT_NO_SCHEME_HEADER, ADF_HTTP2_CLIENT_NO_PATH_HEADER, ADF_HTTP2_CLIENT_PREMATURELY_CLOSED_STREAM, ADF_HTTP2_CLIENT_PREMATURELY_CLOSED_CONNECTION, ADF_HTTP2_CLIENT_LARGER_DATA_BODY_THAN_DECLARED, ADF_HTTP2_CLIENT_LARGE_CHUNKED_BODY, ADF_HTTP2_NEGATIVE_WINDOW_UPDATE, ADF_HTTP2_SEND_WINDOW_FLOW_CONTROL_ERROR, ADF_HTTP2_CLIENT_UNEXPECTED_CONTINUATION_FRAME, ADF_HTTP2_CLIENT_WINDOW_UPDATE_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_WINDOW_UPDATE_FRAME_INCORRECT_INCREMENT, ADF_HTTP2_CLIENT_WINDOW_UPDATE_FRAME_INCREMENT_NOT_ALLOWED_FOR_WINDOW, ADF_HTTP2_CLIENT_GOAWAY_FRAME_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_PING_FRAME_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_PUSH_PROMISE, ADF_HTTP2_CLIENT_SETTINGS_FRAME_INCORRECT_MAX_FRAME_SIZE, ADF_HTTP2_CLIENT_SETTINGS_FRAME_INCORRECT_INIIAL_WINDOW_SIZE, ADF_HTTP2_CLIENT_SETTINGS_FRAME_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_SETTINGS_FRAME_ACK_FLAG_NONZERO_LENGTH, ADF_HTTP2_CLIENT_RST_STREAM_FRAME_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_RST_STREAM_FRAME_INCORRECT_IDENTIFIER, ADF_HTTP2_CLIENT_PRIORITY_FRAME_INCORRECT_DEPENDENCY, ADF_HTTP2_CLIENT_PRIORITY_FRAME_INCORRECT_IDENTIFIER, ADF_HTTP2_CLIENT_PRIORITY_FRAME_INCORRECT_LENGTH, ADF_HTTP2_CLIENT_CONTINUATION_FRAME_INCORRECT_IDENTIFIER, ADF_HTTP2_CLIENT_CONTINUATION_FRAME_EXPECTED_INAPPROPRIATE_FRAME, ADF_HTTP2_CLIENT_INVALID_HEADER, ADF_USER_DELETE_OPERATION_DATASCRIPT_RESET_CONN, ADF_USER_DELETE_OPERATION_HTTP_RULE_SECURITY_ACTION_CLOSE_CONN, ADF_USER_DELETE_OPERATION_HTTP_RULE_SECURITY_RATE_LIMIT_ACTION_CLOSE_CONN, ADF_USER_DELETE_OPERATION_HTTP_RULE_MISSING_TOKEN_ACTION_CLOSE_CONN, ADF_HTTP_BAD_REQUEST_INVALID_HOST_IN_REQUEST_LINE, ADF_HTTP_BAD_REQUEST_RECEIVED_VERSION_LESS_THAN_10, ADF_HTTP_NOT_ALLOWED_DATASCRIPT_RESPONSE_RETURNED_4XX, ADF_HTTP_NOT_ALLOWED_RUM_FLAGGED_INVALID_METHOD, ADF_HTTP_NOT_ALLOWED_UNSUPPORTED_TRACE_METHOD, ADF_HTTP_REQUEST_TIMEOUT_WAITING_FOR_CLIENT, ADF_HTTP_BAD_REQUEST_CLIENT_SENT_INVALID_CONTENT_LENGTH, ADF_HTTP_BAD_REQUEST_CLIENT_SENT_HTTP11_WITHOUT_HOST_HDR, ADF_HTTP_BAD_REQUEST_FAILED_TO_PARSE_URI, ADF_HTTP_BAD_REQUEST_INVALID_HEADER_LINE, ADF_HTTP_BAD_REQUEST_ERROR_WHILE_READING_CLIENT_HEADERS, ADF_HTTP_BAD_REQUEST_CLIENT_SENT_DUPLICATE_HEADER, ADF_HTTP_BAD_REQUEST_CLIENT_SENT_INVALID_HOST_HEADER, ADF_HTTP_NOT_IMPLEMENTED_CLIENT_SENT_UNKNOWN_TRANSFER_ENCODING, ADF_HTTP_BAD_REQUEST_REQUESTED_SERVER_NAME_DIFFERS, ADF_HTTP_BAD_REQUEST_CLIENT_SENT_INVALID_CHUNKED_BODY, ADF_HTTP_BAD_REQUEST_INVALID_HEADER_IN_SPDY, ADF_HTTP_BAD_REQUEST_INVALID_HEADER_BLOCK_IN_SPDY, ADF_HTTP_BAD_REQUEST_DATA_ERROR_IN_SPDY, ADF_HTTP_BAD_REQUEST_NO_METHOD_URI_OR_PROT_IN_REQ_CREATE_SPDY, ADF_HTTP_BAD_REQUEST_CLIENT_PREMATURELY_CLOSED_SPDY_STREAM, ADF_HTTP_BAD_REQUEST_DATA_ERROR_IN_SPDY_READ_REQ_BODY, ADF_HTTP_BAD_REQUEST_CERT_ERROR, ADF_HTTP_BAD_REQUEST_PLAIN_HTTP_REQUEST_SENT_ON_HTTPS_PORT, ADF_HTTP_BAD_REQUEST_NO_CERT_ERROR, ADF_HTTP_BAD_REQUEST_HEADER_TOO_LARGE, ADF_SERVER_HIGH_RESPONSE_TIME_L7, ADF_SERVER_HIGH_RESPONSE_TIME_L4, ADF_COOKIE_SIZE_GREATER_THAN_MAX, ADF_COOKIE_SIZE_LESS_THAN_MIN_COOKIE_LEN, ADF_PERSISTENCE_PROFILE_KEYS_NOT_CONFIGURED, ADF_PERSISTENCE_COOKIE_VERSION_MISMATCH, ADF_COOKIE_ABSENT_FROM_KEYS_IN_PERSISTENCE_PROFILE, ADF_GSLB_SITE_PERSISTENCE_REMOTE_SITE_DOWN, ADF_HTTP_NOT_ALLOWED_DATASCRIPT_RESPONSE_RETURNED_5XX, ADF_SERVER_UPSTREAM_TIMEOUT, ADF_SERVER_UPSTREAM_READ_ERROR, ADF_SERVER_UPSTREAM_RESOLVER_ERROR, ADF_SIP_INVALID_MESSAGE_FROM_CLIENT, ADF_SIP_MESSAGE_UPDATE_FAILED, ADF_SIP_SERVER_UNKNOWN_CALLID, ADF_SIP_REQUEST_FAILED, ADF_SIP_REQUEST_TIMEDOUT, ADF_SIP_CONN_IDLE_TIMEDOUT, ADF_SIP_TRANSACTION_TIMEDOUT, ADF_SIP_SVR_UDP_PORT_NOT_REACHABLE, ADF_SIP_CLT_UDP_PORT_NOT_REACHABLE, ADF_SIP_INVALID_MESSAGE_FROM_SERVER, ADF_SAML_COOKIE_VERSION_MISMATCH, ADF_SAML_COOKIE_KEYS_NOT_CONFIGURED, ADF_SAML_COOKIE_ABSENT_FROM_KEYS_IN_SAML_AUTH_POLICY, ADF_SAML_COOKIE_INVALID, ADF_SAML_COOKIE_DECRYPTION_ERROR, ADF_SAML_COOKIE_ENCRYPTION_ERROR, ADF_SAML_COOKIE_DECODE_ERROR, ADF_SAML_COOKIE_SESSION_COOKIE_GREATER_THAN_MAX, ADF_SAML_ASSERTION_DOES_NOT_MATCH_REQUEST_ID.
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

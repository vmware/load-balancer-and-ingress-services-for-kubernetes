package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VserverL7MetricsObj vserver l7 metrics obj
// swagger:model VserverL7MetricsObj
type VserverL7MetricsObj struct {

	// Client Apdex measures quality of server response based on latency.
	Apdexr *float64 `json:"apdexr,omitempty"`

	// Average server/application response latency.
	AvgApplicationResponseTime *float64 `json:"avg_application_response_time,omitempty"`

	// Average time client was blocked as reported by client.
	AvgBlockingTime *float64 `json:"avg_blocking_time,omitempty"`

	// Average browser rendering latency.
	AvgBrowserRenderingTime *float64 `json:"avg_browser_rendering_time,omitempty"`

	// Average cache bytes.
	AvgCacheBytes *float64 `json:"avg_cache_bytes,omitempty"`

	// Average cache hit of requests.
	AvgCacheHits *float64 `json:"avg_cache_hits,omitempty"`

	// Average cacheable bytes.
	AvgCacheableBytes *float64 `json:"avg_cacheable_bytes,omitempty"`

	// Average cacheable hit of requests.
	AvgCacheableHits *float64 `json:"avg_cacheable_hits,omitempty"`

	// Average client data transfer time that represents latency of sending response to the client excluding the RTT time . Higher client data transfer time signifies lower bandwidth  between client and Avi Service Engine.
	AvgClientDataTransferTime *float64 `json:"avg_client_data_transfer_time,omitempty"`

	// Average client Round Trip Time.
	AvgClientRtt *float64 `json:"avg_client_rtt,omitempty"`

	// Average client transaction latency computed by adding response latencies across all HTTP requests.
	AvgClientTxnLatency *float64 `json:"avg_client_txn_latency,omitempty"`

	// Rate of HTTP responses sent per second.
	AvgCompleteResponses *float64 `json:"avg_complete_responses,omitempty"`

	// Average client connection latency reported by client.
	AvgConnectionTime *float64 `json:"avg_connection_time,omitempty"`

	// Average domain lookup latency reported by client.
	AvgDNSLookupTime *float64 `json:"avg_dns_lookup_time,omitempty"`

	// Average Dom content Load Time reported by clients.
	AvgDomContentLoadTime *float64 `json:"avg_dom_content_load_time,omitempty"`

	// Rate of HTTP error responses sent per second. It does not include errors excluded in analytics profile.
	AvgErrorResponses *float64 `json:"avg_error_responses,omitempty"`

	// Rate of HTTP responses excluded as errors based on analytics profile.
	AvgErrorsExcluded *float64 `json:"avg_errors_excluded,omitempty"`

	// Avg number of HTTP requests that completed within frustrated latency.
	AvgFrustratedResponses *float64 `json:"avg_frustrated_responses,omitempty"`

	// Average size of HTTP headers per request. Field introduced in 17.2.12, 18.1.2.
	AvgHTTPHeadersBytes *float64 `json:"avg_http_headers_bytes,omitempty"`

	// Average number of HTTP headers per request. Field introduced in 17.2.12, 18.1.2.
	AvgHTTPHeadersCount *float64 `json:"avg_http_headers_count,omitempty"`

	// Average number of HTTP request parameters per request. Field introduced in 17.2.12, 18.1.2.
	AvgHTTPParamsCount *float64 `json:"avg_http_params_count,omitempty"`

	// Average Page Load time reported by clients.
	AvgPageDownloadTime *float64 `json:"avg_page_download_time,omitempty"`

	// Average Page Load Time reported by client.
	AvgPageLoadTime *float64 `json:"avg_page_load_time,omitempty"`

	// Average number of HTTP request parameters per request, taking into account only requests with parameters. Field introduced in 17.2.12, 18.1.3.
	AvgParamsPerReq *float64 `json:"avg_params_per_req,omitempty"`

	// Average size of HTTP POST request. Field introduced in 17.2.12, 18.1.2.
	AvgPostBytes *float64 `json:"avg_post_bytes,omitempty"`

	// Average post compression bytes.
	AvgPostCompressionBytes *float64 `json:"avg_post_compression_bytes,omitempty"`

	// Average pre compression bytes.
	AvgPreCompressionBytes *float64 `json:"avg_pre_compression_bytes,omitempty"`

	// Average redirect latency reported by client.
	AvgRedirectionTime *float64 `json:"avg_redirection_time,omitempty"`

	// Average requests per session measured for closed sessions.
	AvgReqsPerSession *float64 `json:"avg_reqs_per_session,omitempty"`

	// Rate of 1xx HTTP responses sent per second.
	AvgResp1xx *float64 `json:"avg_resp_1xx,omitempty"`

	// Rate of 2xx HTTP responses sent per second.
	AvgResp2xx *float64 `json:"avg_resp_2xx,omitempty"`

	// Rate of 3xx HTTP responses sent per second.
	AvgResp3xx *float64 `json:"avg_resp_3xx,omitempty"`

	// Rate of 4xx HTTP responses sent per second.
	AvgResp4xx *float64 `json:"avg_resp_4xx,omitempty"`

	// Rate of 4xx HTTP responses as errors sent by avi. It does not include any error codes excluded in the analytics profile and pool server errors.
	AvgResp4xxAviErrors *float64 `json:"avg_resp_4xx_avi_errors,omitempty"`

	// Rate of 5xx HTTP responses sent per second.
	AvgResp5xx *float64 `json:"avg_resp_5xx,omitempty"`

	// Rate of 5xx HTTP responses as errors sent by avi. It does not include any error codes excluded in the analytics profile and pool server errors.
	AvgResp5xxAviErrors *float64 `json:"avg_resp_5xx_avi_errors,omitempty"`

	// Total client data transfer time by client.
	AvgRumClientDataTransferTime *float64 `json:"avg_rum_client_data_transfer_time,omitempty"`

	// Avg number of HTTP requests that completed within satisfactory latency.
	AvgSatisfactoryResponses *float64 `json:"avg_satisfactory_responses,omitempty"`

	// Average server Round Trip Time.
	AvgServerRtt *float64 `json:"avg_server_rtt,omitempty"`

	// Average latency from receipt of request to start of response.
	AvgServiceTime *float64 `json:"avg_service_time,omitempty"`

	// Average SSL Sessions using DSA certificate.
	AvgSslAuthDsa *float64 `json:"avg_ssl_auth_dsa,omitempty"`

	// Average SSL Sessions using Elliptic Curve DSA (ECDSA) certificates.
	AvgSslAuthEcdsa *float64 `json:"avg_ssl_auth_ecdsa,omitempty"`

	// Average SSL Sessions using RSA certificate.
	AvgSslAuthRsa *float64 `json:"avg_ssl_auth_rsa,omitempty"`

	// Average SSL Sessions.
	AvgSslConnections *float64 `json:"avg_ssl_connections,omitempty"`

	// Average SSL Exchanges using EC Cerificates without PFS.
	AvgSslEcdsaNonPfs *float64 `json:"avg_ssl_ecdsa_non_pfs,omitempty"`

	// Average SSL Exchanges using EC Cerificates and PFS.
	AvgSslEcdsaPfs *float64 `json:"avg_ssl_ecdsa_pfs,omitempty"`

	// Average SSL errors due to clients, protocol errors,network errors and handshake timeouts.
	AvgSslErrors *float64 `json:"avg_ssl_errors,omitempty"`

	// Average SSL connections failed due to protocol , network or timeout reasons.
	AvgSslFailedConnections *float64 `json:"avg_ssl_failed_connections,omitempty"`

	// Average SSL handshakes failed due to network errors.
	AvgSslHandshakeNetworkErrors *float64 `json:"avg_ssl_handshake_network_errors,omitempty"`

	// Average SSL handshake failed due to clients or protocol errors.
	AvgSslHandshakeProtocolErrors *float64 `json:"avg_ssl_handshake_protocol_errors,omitempty"`

	// Average new successful SSL sessions.
	AvgSslHandshakesNew *float64 `json:"avg_ssl_handshakes_new,omitempty"`

	// Average SSL Exchanges using Non-PFS.
	AvgSslHandshakesNonPfs *float64 `json:"avg_ssl_handshakes_non_pfs,omitempty"`

	// Average SSL Exchanges using PFS.
	AvgSslHandshakesPfs *float64 `json:"avg_ssl_handshakes_pfs,omitempty"`

	// Average new successful resumed SSL sessions.
	AvgSslHandshakesReused *float64 `json:"avg_ssl_handshakes_reused,omitempty"`

	// Average SSL handshakes timed out.
	AvgSslHandshakesTimedout *float64 `json:"avg_ssl_handshakes_timedout,omitempty"`

	// Average SSL Exchanges using Diffie-Hellman.
	AvgSslKxDh *float64 `json:"avg_ssl_kx_dh,omitempty"`

	// Average SSL Exchanges using RSA.
	AvgSslKxEcdh *float64 `json:"avg_ssl_kx_ecdh,omitempty"`

	// Average SSL Exchanges using RSA.
	AvgSslKxRsa *float64 `json:"avg_ssl_kx_rsa,omitempty"`

	// Average SSL Exchanges using RSA Cerificates without PFS.
	AvgSslRsaNonPfs *float64 `json:"avg_ssl_rsa_non_pfs,omitempty"`

	// Average SSL Exchanges using RSA Cerificates and PFS.
	AvgSslRsaPfs *float64 `json:"avg_ssl_rsa_pfs,omitempty"`

	// Average SSL Sessions with version 3.0.
	AvgSslVerSsl30 *float64 `json:"avg_ssl_ver_ssl30,omitempty"`

	// Average SSL Sessions with TLS version 1.0.
	AvgSslVerTLS10 *float64 `json:"avg_ssl_ver_tls10,omitempty"`

	// Average SSL Sessions with TLS version 1.1.
	AvgSslVerTLS11 *float64 `json:"avg_ssl_ver_tls11,omitempty"`

	// Average SSL Sessions with TLS version 1.2.
	AvgSslVerTLS12 *float64 `json:"avg_ssl_ver_tls12,omitempty"`

	// Avg number of HTTP requests that completed within tolerated latency.
	AvgToleratedResponses *float64 `json:"avg_tolerated_responses,omitempty"`

	// Average rate of client HTTP requests received by the virtual service per second.
	AvgTotalRequests *float64 `json:"avg_total_requests,omitempty"`

	// Average length of HTTP URI per request. Field introduced in 17.2.12, 18.1.2.
	AvgURILength *float64 `json:"avg_uri_length,omitempty"`

	// Average number of transactions per second identified by WAF as attacks. Field introduced in 17.2.3.
	AvgWafAttacks *float64 `json:"avg_waf_attacks,omitempty"`

	// Average number of transactions per second bypassing WAF. Field introduced in 17.2.12, 18.1.2.
	AvgWafDisabled *float64 `json:"avg_waf_disabled,omitempty"`

	// Average number of transactions per second evaluated by WAF. Field introduced in 17.2.2.
	AvgWafEvaluated *float64 `json:"avg_waf_evaluated,omitempty"`

	// Average number of requests per second evaluated by WAF in Request Body Phase. Field introduced in 17.2.2.
	AvgWafEvaluatedRequestBodyPhase *float64 `json:"avg_waf_evaluated_request_body_phase,omitempty"`

	// Average number of requests per second evaluated by WAF in Request Header Phase. Field introduced in 17.2.2.
	AvgWafEvaluatedRequestHeaderPhase *float64 `json:"avg_waf_evaluated_request_header_phase,omitempty"`

	// Average number of responses per second evaluated by WAF in Response Body Phase. Field introduced in 17.2.2.
	AvgWafEvaluatedResponseBodyPhase *float64 `json:"avg_waf_evaluated_response_body_phase,omitempty"`

	// Average number of responsess per second evaluated by WAF in Response Header Phase. Field introduced in 17.2.2.
	AvgWafEvaluatedResponseHeaderPhase *float64 `json:"avg_waf_evaluated_response_header_phase,omitempty"`

	// Average number of transactions per second flagged by WAF. Field introduced in 17.2.2.
	AvgWafFlagged *float64 `json:"avg_waf_flagged,omitempty"`

	// Average number of requests per second flagged (but not rejected) by WAF in Request Body Phase. Field introduced in 17.2.2.
	AvgWafFlaggedRequestBodyPhase *float64 `json:"avg_waf_flagged_request_body_phase,omitempty"`

	// Average number of requests per second flagged (but not rejected) by WAF in Request Header Phase. Field introduced in 17.2.2.
	AvgWafFlaggedRequestHeaderPhase *float64 `json:"avg_waf_flagged_request_header_phase,omitempty"`

	// Average number of responses per second flagged (but not rejected) by WAF in Response Body Phase. Field introduced in 17.2.2.
	AvgWafFlaggedResponseBodyPhase *float64 `json:"avg_waf_flagged_response_body_phase,omitempty"`

	// Average number of responses per second flagged (but not rejected) by WAF in Response Header Phase. Field introduced in 17.2.2.
	AvgWafFlaggedResponseHeaderPhase *float64 `json:"avg_waf_flagged_response_header_phase,omitempty"`

	// Average waf latency seen due to WAF Request Body processing. Field introduced in 17.2.2.
	AvgWafLatencyRequestBodyPhase *float64 `json:"avg_waf_latency_request_body_phase,omitempty"`

	// Average waf latency seen due to WAF Request Header processing. Field introduced in 17.2.2.
	AvgWafLatencyRequestHeaderPhase *float64 `json:"avg_waf_latency_request_header_phase,omitempty"`

	// Average waf latency seen due to WAF Response Body processing. Field introduced in 17.2.2.
	AvgWafLatencyResponseBodyPhase *float64 `json:"avg_waf_latency_response_body_phase,omitempty"`

	// Average waf latency seen due to WAF Response Header processing. Field introduced in 17.2.2.
	AvgWafLatencyResponseHeaderPhase *float64 `json:"avg_waf_latency_response_header_phase,omitempty"`

	// Average number of transactions per second matched by WAF rule/rules. Field introduced in 17.2.2.
	AvgWafMatched *float64 `json:"avg_waf_matched,omitempty"`

	// Average number of requests per second matched by WAF in Request Body Phase. Field introduced in 17.2.2.
	AvgWafMatchedRequestBodyPhase *float64 `json:"avg_waf_matched_request_body_phase,omitempty"`

	// Average number of requests per second matched by WAF in Request Header Phase. Field introduced in 17.2.2.
	AvgWafMatchedRequestHeaderPhase *float64 `json:"avg_waf_matched_request_header_phase,omitempty"`

	// Average number of responses per second matched by WAF in Response Body Phase. Field introduced in 17.2.2.
	AvgWafMatchedResponseBodyPhase *float64 `json:"avg_waf_matched_response_body_phase,omitempty"`

	// Average number of responses per second matched by WAF in Response Header Phase. Field introduced in 17.2.2.
	AvgWafMatchedResponseHeaderPhase *float64 `json:"avg_waf_matched_response_header_phase,omitempty"`

	// Average number of transactions per second rejected by WAF. Field introduced in 17.2.2.
	AvgWafRejected *float64 `json:"avg_waf_rejected,omitempty"`

	// Average number of requests per second rejected by WAF in Request Body Phase. Field introduced in 17.2.2.
	AvgWafRejectedRequestBodyPhase *float64 `json:"avg_waf_rejected_request_body_phase,omitempty"`

	// Average number of requests per second rejected by WAF in Request Header Phase. Field introduced in 17.2.2.
	AvgWafRejectedRequestHeaderPhase *float64 `json:"avg_waf_rejected_request_header_phase,omitempty"`

	// Average number of responses per second rejected by WAF in Response Body Phase. Field introduced in 17.2.2.
	AvgWafRejectedResponseBodyPhase *float64 `json:"avg_waf_rejected_response_body_phase,omitempty"`

	// Average number of responses per second rejected by WAF in Response Header Phase. Field introduced in 17.2.2.
	AvgWafRejectedResponseHeaderPhase *float64 `json:"avg_waf_rejected_response_header_phase,omitempty"`

	// Average Waiting Time reported by client.
	AvgWaitingTime *float64 `json:"avg_waiting_time,omitempty"`

	// Maximum number of concurrent HTTP sessions.
	MaxConcurrentSessions *float64 `json:"max_concurrent_sessions,omitempty"`

	// Maximum number of open SSL sessions.
	MaxSslOpenSessions *float64 `json:"max_ssl_open_sessions,omitempty"`

	// node_obj_id of VserverL7MetricsObj.
	// Required: true
	NodeObjID *string `json:"node_obj_id"`

	// Percentage cache hit of requests.
	PctCacheHits *float64 `json:"pct_cache_hits,omitempty"`

	// Percentage cacheable hit of requests.
	PctCacheableHits *float64 `json:"pct_cacheable_hits,omitempty"`

	// Number of HTTP GET requests as a percentage of total requests received. Field introduced in 17.2.12, 18.1.2.
	PctGetReqs *float64 `json:"pct_get_reqs,omitempty"`

	// Number of HTTP POST requests as a percentage of total requests received. Field introduced in 17.2.12, 18.1.2.
	PctPostReqs *float64 `json:"pct_post_reqs,omitempty"`

	// Percent of 4xx and 5xx responses.
	PctResponseErrors *float64 `json:"pct_response_errors,omitempty"`

	// Percent of SSL connections failured due to protocol , network or timeout reasons.
	PctSslFailedConnections *float64 `json:"pct_ssl_failed_connections,omitempty"`

	// Malicious transactions (Attacks) identified by WAF as the pecentage  of total requests received. Field introduced in 17.2.3.
	PctWafAttacks *float64 `json:"pct_waf_attacks,omitempty"`

	// Transactions bypassing WAF as the percentage of total requests received. Field introduced in 17.2.12, 18.1.2.
	PctWafDisabled *float64 `json:"pct_waf_disabled,omitempty"`

	// WAF evaluated transactions as the pecentage of total requests received. Field introduced in 17.2.2.
	PctWafEvaluated *float64 `json:"pct_waf_evaluated,omitempty"`

	// WAF flagged transactions as the percentage of total WAF evaluated transactions. Field introduced in 17.2.2.
	PctWafFlagged *float64 `json:"pct_waf_flagged,omitempty"`

	// WAF matched requests as the percentage of total WAF evaluated requests. Field introduced in 17.2.2.
	PctWafMatched *float64 `json:"pct_waf_matched,omitempty"`

	// WAF rejected transactions as the percentage of total WAF evaluated transactions. Field introduced in 17.2.2.
	PctWafRejected *float64 `json:"pct_waf_rejected,omitempty"`

	// Apdex measures quality of server response based on Real User Metric.
	RumApdexr *float64 `json:"rum_apdexr,omitempty"`

	// Protocol strength of SSL ciphers used.
	SslProtocolStrength *float64 `json:"ssl_protocol_strength,omitempty"`

	// Total time taken by server to respond to requesti.
	SumApplicationResponseTime *float64 `json:"sum_application_response_time,omitempty"`

	// Total time client was blocked.
	SumBlockingTime *float64 `json:"sum_blocking_time,omitempty"`

	// Total browser rendering latency reported by client.
	SumBrowserRenderingTime *float64 `json:"sum_browser_rendering_time,omitempty"`

	// Average client data transfer time computed by adding response latencies across all HTTP requests.
	SumClientDataTransferTime *float64 `json:"sum_client_data_transfer_time,omitempty"`

	// Sum of all client Round Trip Times for all samples.
	SumClientRtt *float64 `json:"sum_client_rtt,omitempty"`

	// Total client connection latency reported by client.
	SumConnectionTime *float64 `json:"sum_connection_time,omitempty"`

	// Total domain lookup latency reported by client.
	SumDNSLookupTime *float64 `json:"sum_dns_lookup_time,omitempty"`

	// Total dom content latency reported by client.
	SumDomContentLoadTime *float64 `json:"sum_dom_content_load_time,omitempty"`

	// Count of HTTP 400 and 500 errors for a virtual service in a time interval.
	SumErrors *float64 `json:"sum_errors,omitempty"`

	// Number of server sessions closed in this interval.
	SumFinishedSessions *float64 `json:"sum_finished_sessions,omitempty"`

	// Total latency from responses to all the GET requests.
	SumGetClientTxnLatency *float64 `json:"sum_get_client_txn_latency,omitempty"`

	// Total number of HTTP GET requests that were responded satisfactorily within latency threshold.
	SumGetClientTxnLatencyBucket1 *float64 `json:"sum_get_client_txn_latency_bucket1,omitempty"`

	// Total number of HTTP GET requests that were responded beyond latency threshold but within tolerated limits.
	SumGetClientTxnLatencyBucket2 *float64 `json:"sum_get_client_txn_latency_bucket2,omitempty"`

	// Total number of HTTP GET requests.
	SumGetReqs *float64 `json:"sum_get_reqs,omitempty"`

	// Total size of HTTP request headers. Field introduced in 17.2.12, 18.1.2.
	SumHTTPHeadersBytes *float64 `json:"sum_http_headers_bytes,omitempty"`

	// Total number of HTTP headers across all requests in a given metrics interval. Field introduced in 17.2.12, 18.1.2.
	SumHTTPHeadersCount *float64 `json:"sum_http_headers_count,omitempty"`

	// Total number of HTTP request parameters. Field introduced in 17.2.12, 18.1.2.
	SumHTTPParamsCount *float64 `json:"sum_http_params_count,omitempty"`

	// Total samples that had satisfactory page load time.
	SumNumPageLoadTimeBucket1 *float64 `json:"sum_num_page_load_time_bucket1,omitempty"`

	// Total samples that had tolerated page load time.
	SumNumPageLoadTimeBucket2 *float64 `json:"sum_num_page_load_time_bucket2,omitempty"`

	// Total samples used for rum metrics.
	SumNumRumSamples *float64 `json:"sum_num_rum_samples,omitempty"`

	// Total latency from responses to all the requests other than GET or POST.
	SumOtherClientTxnLatency *float64 `json:"sum_other_client_txn_latency,omitempty"`

	// Total number of HTTP requests other than GET or POST that were responded satisfactorily within latency threshold.
	SumOtherClientTxnLatencyBucket1 *float64 `json:"sum_other_client_txn_latency_bucket1,omitempty"`

	// Total number of HTTP requests other than GET or POST that were responded beyond latency threshold but within tolerated limits.
	SumOtherClientTxnLatencyBucket2 *float64 `json:"sum_other_client_txn_latency_bucket2,omitempty"`

	// Total number of HTTP requests that are not GET or POST requests.
	SumOtherReqs *float64 `json:"sum_other_reqs,omitempty"`

	// Total time to transfer response to client.
	SumPageDownloadTime *float64 `json:"sum_page_download_time,omitempty"`

	// Total Page Load Time reported by client.
	SumPageLoadTime *float64 `json:"sum_page_load_time,omitempty"`

	// Total size of HTTP POST requests. Field introduced in 17.2.12, 18.1.2.
	SumPostBytes *float64 `json:"sum_post_bytes,omitempty"`

	// Total latency from responses to all the POST requests.
	SumPostClientTxnLatency *float64 `json:"sum_post_client_txn_latency,omitempty"`

	// Total number of HTTP POST requests that were responded satisfactorily within latency threshold.
	SumPostClientTxnLatencyBucket1 *float64 `json:"sum_post_client_txn_latency_bucket1,omitempty"`

	// Total number of HTTP POST requests that were responded beyond latency threshold but within tolerated limits.
	SumPostClientTxnLatencyBucket2 *float64 `json:"sum_post_client_txn_latency_bucket2,omitempty"`

	// Total number of HTTP POST requests.
	SumPostReqs *float64 `json:"sum_post_reqs,omitempty"`

	// Total redirect latency reported by client.
	SumRedirectionTime *float64 `json:"sum_redirection_time,omitempty"`

	// Total number of requests served across server sessions closed in the interval.
	SumReqsFinishedSessions *float64 `json:"sum_reqs_finished_sessions,omitempty"`

	// Total number of HTTP requests containing at least one parameter. Field introduced in 17.2.12, 18.1.3.
	SumReqsWithParams *float64 `json:"sum_reqs_with_params,omitempty"`

	// Total number of HTTP 1XX responses.
	SumResp1xx *float64 `json:"sum_resp_1xx,omitempty"`

	// Total number of HTTP 2XX responses.
	SumResp2xx *float64 `json:"sum_resp_2xx,omitempty"`

	// Total number of HTTP 3XX responses.
	SumResp3xx *float64 `json:"sum_resp_3xx,omitempty"`

	// Total number of HTTP 4XX error responses.
	SumResp4xx *float64 `json:"sum_resp_4xx,omitempty"`

	// Total number of HTTP 5XX error responses.
	SumResp5xx *float64 `json:"sum_resp_5xx,omitempty"`

	// Total client data transfer time by client.
	SumRumClientDataTransferTime *float64 `json:"sum_rum_client_data_transfer_time,omitempty"`

	// Sum of all server Round Trip Times for all samples.
	SumServerRtt *float64 `json:"sum_server_rtt,omitempty"`

	// Total time from receipt of request to start of response.
	SumServiceTime *float64 `json:"sum_service_time,omitempty"`

	// Total number of HTTP responses sent.
	SumTotalResponses *float64 `json:"sum_total_responses,omitempty"`

	// Total length of HTTP request URIs. Field introduced in 17.2.12, 18.1.2.
	SumURILength *float64 `json:"sum_uri_length,omitempty"`

	// Total number of transactions identified by WAF as attacks. Field introduced in 17.2.3.
	SumWafAttacks *float64 `json:"sum_waf_attacks,omitempty"`

	// Total number of requests bypassing WAF. Field introduced in 17.2.12, 18.1.2.
	SumWafDisabled *float64 `json:"sum_waf_disabled,omitempty"`

	// Total number of requests evaluated by WAF in Request Body Phase. Field introduced in 17.2.2.
	SumWafEvaluatedRequestBodyPhase *float64 `json:"sum_waf_evaluated_request_body_phase,omitempty"`

	// Total number of requests evaluated by WAF in Request Header Phase. Field introduced in 17.2.2.
	SumWafEvaluatedRequestHeaderPhase *float64 `json:"sum_waf_evaluated_request_header_phase,omitempty"`

	// Total number of responses evaluated by WAF in Response Body Phase. Field introduced in 17.2.2.
	SumWafEvaluatedResponseBodyPhase *float64 `json:"sum_waf_evaluated_response_body_phase,omitempty"`

	// Total number of responses evaluated by WAF in Response Header Phase. Field introduced in 17.2.2.
	SumWafEvaluatedResponseHeaderPhase *float64 `json:"sum_waf_evaluated_response_header_phase,omitempty"`

	// Total number of transactions (requests or responses) flagged as attack by WAF. Field introduced in 17.2.3.
	SumWafFlagged *float64 `json:"sum_waf_flagged,omitempty"`

	// Total number of requests flagged (but not rejected) by WAF in Request Body Phase. Field introduced in 17.2.2.
	SumWafFlaggedRequestBodyPhase *float64 `json:"sum_waf_flagged_request_body_phase,omitempty"`

	// Total number of requests flagged (but not rejected) by WAF in Request Header Phase. Field introduced in 17.2.2.
	SumWafFlaggedRequestHeaderPhase *float64 `json:"sum_waf_flagged_request_header_phase,omitempty"`

	// Total number of responses flagged (but not rejected) by WAF in Response Body Phase. Field introduced in 17.2.2.
	SumWafFlaggedResponseBodyPhase *float64 `json:"sum_waf_flagged_response_body_phase,omitempty"`

	// Total number of responses flagged (but not rejected) by WAF in Response Header Phase. Field introduced in 17.2.2.
	SumWafFlaggedResponseHeaderPhase *float64 `json:"sum_waf_flagged_response_header_phase,omitempty"`

	// Total latency seen by all evaluated requests in Request Body Phase. Field introduced in 17.2.2.
	SumWafLatencyRequestBodyPhase *float64 `json:"sum_waf_latency_request_body_phase,omitempty"`

	// Total latency seen by all transactions evaluated by WAF in Request Header Phase. Field introduced in 17.2.2.
	SumWafLatencyRequestHeaderPhase *float64 `json:"sum_waf_latency_request_header_phase,omitempty"`

	// Total latency seen by all evaluated responses in Response Body Phase. Field introduced in 17.2.2.
	SumWafLatencyResponseBodyPhase *float64 `json:"sum_waf_latency_response_body_phase,omitempty"`

	// Total latency seen by all evaluated responsess in WAF Response Header Phase. Field introduced in 17.2.2.
	SumWafLatencyResponseHeaderPhase *float64 `json:"sum_waf_latency_response_header_phase,omitempty"`

	// Total number of requests matched by WAF in Request Body Phase. Field introduced in 17.2.2.
	SumWafMatchedRequestBodyPhase *float64 `json:"sum_waf_matched_request_body_phase,omitempty"`

	// Total number of requests matched by WAF in Request Header Phase. Field introduced in 17.2.2.
	SumWafMatchedRequestHeaderPhase *float64 `json:"sum_waf_matched_request_header_phase,omitempty"`

	// Total number of responses matched by WAF in Response Body Phase. Field introduced in 17.2.2.
	SumWafMatchedResponseBodyPhase *float64 `json:"sum_waf_matched_response_body_phase,omitempty"`

	// Total number of responses matched by WAF in Response Header Phase. Field introduced in 17.2.2.
	SumWafMatchedResponseHeaderPhase *float64 `json:"sum_waf_matched_response_header_phase,omitempty"`

	// Total number of transactions (requests or responses) rejected by WAF. Field introduced in 17.2.3.
	SumWafRejected *float64 `json:"sum_waf_rejected,omitempty"`

	// Total number of requests rejected by WAF in Request Body Phase. Field introduced in 17.2.2.
	SumWafRejectedRequestBodyPhase *float64 `json:"sum_waf_rejected_request_body_phase,omitempty"`

	// Total number of requests rejected by WAF in Request Header Phase. Field introduced in 17.2.2.
	SumWafRejectedRequestHeaderPhase *float64 `json:"sum_waf_rejected_request_header_phase,omitempty"`

	// Total number of responses rejected by WAF in Response Body Phase. Field introduced in 17.2.2.
	SumWafRejectedResponseBodyPhase *float64 `json:"sum_waf_rejected_response_body_phase,omitempty"`

	// Total number of responses rejected by WAF in Response Header Phase. Field introduced in 17.2.2.
	SumWafRejectedResponseHeaderPhase *float64 `json:"sum_waf_rejected_response_header_phase,omitempty"`

	// Total waiting reported by client.
	SumWaitingTime *float64 `json:"sum_waiting_time,omitempty"`
}

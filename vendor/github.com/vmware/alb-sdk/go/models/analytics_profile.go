// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AnalyticsProfile analytics profile
// swagger:model AnalyticsProfile
type AnalyticsProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// If a client receives an HTTP response in less than the Satisfactory Latency Threshold, the request is considered Satisfied. It is considered Tolerated if it is not Satisfied and less than Tolerated Latency Factor multiplied by the Satisfactory Latency Threshold. Greater than this number and the client's request is considered Frustrated. Allowed values are 1-30000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 500), Basic edition(Allowed values- 500), Enterprise with Cloud Services edition.
	ApdexResponseThreshold *uint32 `json:"apdex_response_threshold,omitempty"`

	// Client tolerated response latency factor. Client must receive a response within this factor times the satisfactory threshold (apdex_response_threshold) to be considered tolerated. Allowed values are 1-1000. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 4), Basic edition(Allowed values- 4), Enterprise with Cloud Services edition.
	ApdexResponseToleratedFactor *float64 `json:"apdex_response_tolerated_factor,omitempty"`

	// Satisfactory client to Avi Round Trip Time(RTT). Allowed values are 1-2000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 250), Basic edition(Allowed values- 250), Enterprise with Cloud Services edition.
	ApdexRttThreshold *uint32 `json:"apdex_rtt_threshold,omitempty"`

	// Tolerated client to Avi Round Trip Time(RTT) factor.  It is a multiple of apdex_rtt_tolerated_factor. Allowed values are 1-1000. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 4), Basic edition(Allowed values- 4), Enterprise with Cloud Services edition.
	ApdexRttToleratedFactor *float64 `json:"apdex_rtt_tolerated_factor,omitempty"`

	// If a client is able to load a page in less than the Satisfactory Latency Threshold, the PageLoad is considered Satisfied.  It is considered tolerated if it is greater than Satisfied but less than the Tolerated Latency multiplied by Satisifed Latency. Greater than this number and the client's request is considered Frustrated.  A PageLoad includes the time for DNS lookup, download of all HTTP objects, and page render time. Allowed values are 1-30000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5000), Basic edition(Allowed values- 5000), Enterprise with Cloud Services edition.
	ApdexRumThreshold *uint32 `json:"apdex_rum_threshold,omitempty"`

	// Virtual service threshold factor for tolerated Page Load Time (PLT) as multiple of apdex_rum_threshold. Allowed values are 1-1000. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 4), Basic edition(Allowed values- 4), Enterprise with Cloud Services edition.
	ApdexRumToleratedFactor *float64 `json:"apdex_rum_tolerated_factor,omitempty"`

	// A server HTTP response is considered Satisfied if latency is less than the Satisfactory Latency Threshold. The response is considered tolerated when it is greater than Satisfied but less than the Tolerated Latency Factor * S_Latency.  Greater than this number and the server response is considered Frustrated. Allowed values are 1-30000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 400), Basic edition(Allowed values- 400), Enterprise with Cloud Services edition.
	ApdexServerResponseThreshold *uint32 `json:"apdex_server_response_threshold,omitempty"`

	// Server tolerated response latency factor. Servermust response within this factor times the satisfactory threshold (apdex_server_response_threshold) to be considered tolerated. Allowed values are 1-1000. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 4), Basic edition(Allowed values- 4), Enterprise with Cloud Services edition.
	ApdexServerResponseToleratedFactor *float64 `json:"apdex_server_response_tolerated_factor,omitempty"`

	// Satisfactory client to Avi Round Trip Time(RTT). Allowed values are 1-2000. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 125), Basic edition(Allowed values- 125), Enterprise with Cloud Services edition.
	ApdexServerRttThreshold *uint32 `json:"apdex_server_rtt_threshold,omitempty"`

	// Tolerated client to Avi Round Trip Time(RTT) factor.  It is a multiple of apdex_rtt_tolerated_factor. Allowed values are 1-1000. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 4), Basic edition(Allowed values- 4), Enterprise with Cloud Services edition.
	ApdexServerRttToleratedFactor *float64 `json:"apdex_server_rtt_tolerated_factor,omitempty"`

	// Configure which logs are sent to the Avi Controller from SEs and how they are processed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientLogConfig *ClientLogConfiguration `json:"client_log_config,omitempty"`

	// Configure to stream logs to an external server. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientLogStreamingConfig *ClientLogStreamingConfig `json:"client_log_streaming_config,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// A connection between client and Avi is considered lossy when more than this percentage of out of order packets are received. Allowed values are 1-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 50), Basic edition(Allowed values- 50), Enterprise with Cloud Services edition.
	ConnLossyOooThreshold *uint32 `json:"conn_lossy_ooo_threshold,omitempty"`

	// A connection between client and Avi is considered lossy when more than this percentage of packets are retransmitted due to timeout. Allowed values are 1-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 20), Basic edition(Allowed values- 20), Enterprise with Cloud Services edition.
	ConnLossyTimeoRexmtThreshold *uint32 `json:"conn_lossy_timeo_rexmt_threshold,omitempty"`

	// A connection between client and Avi is considered lossy when more than this percentage of packets are retransmitted. Allowed values are 1-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 50), Basic edition(Allowed values- 50), Enterprise with Cloud Services edition.
	ConnLossyTotalRexmtThreshold *uint32 `json:"conn_lossy_total_rexmt_threshold,omitempty"`

	// A client connection is considered lossy when percentage of times a packet could not be trasmitted due to TCP zero window is above this threshold. Allowed values are 0-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 2), Basic edition(Allowed values- 2), Enterprise with Cloud Services edition.
	ConnLossyZeroWinSizeEventThreshold *uint32 `json:"conn_lossy_zero_win_size_event_threshold,omitempty"`

	// A connection between Avi and server is considered lossy when more than this percentage of out of order packets are received. Allowed values are 1-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 50), Basic edition(Allowed values- 50), Enterprise with Cloud Services edition.
	ConnServerLossyOooThreshold *uint32 `json:"conn_server_lossy_ooo_threshold,omitempty"`

	// A connection between Avi and server is considered lossy when more than this percentage of packets are retransmitted due to timeout. Allowed values are 1-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 20), Basic edition(Allowed values- 20), Enterprise with Cloud Services edition.
	ConnServerLossyTimeoRexmtThreshold *uint32 `json:"conn_server_lossy_timeo_rexmt_threshold,omitempty"`

	// A connection between Avi and server is considered lossy when more than this percentage of packets are retransmitted. Allowed values are 1-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 50), Basic edition(Allowed values- 50), Enterprise with Cloud Services edition.
	ConnServerLossyTotalRexmtThreshold *uint32 `json:"conn_server_lossy_total_rexmt_threshold,omitempty"`

	// A server connection is considered lossy when percentage of times a packet could not be trasmitted due to TCP zero window is above this threshold. Allowed values are 0-100. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 2), Basic edition(Allowed values- 2), Enterprise with Cloud Services edition.
	ConnServerLossyZeroWinSizeEventThreshold *uint32 `json:"conn_server_lossy_zero_win_size_event_threshold,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable adaptive configuration for optimizing resource usage. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableAdaptiveConfig *bool `json:"enable_adaptive_config,omitempty"`

	// Enables Advanced Analytics features like Anomaly detection. If set to false, anomaly computation (and associated rules/events) for VS, Pool and Server metrics will be deactivated. However, setting it to false reduces cpu and memory requirements for Analytics subsystem. Field introduced in 17.2.13, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Essentials edition is false, Basic edition is false, Enterprise is True.
	EnableAdvancedAnalytics *bool `json:"enable_advanced_analytics,omitempty"`

	// Virtual Service (VS) metrics are processed only when there is live data traffic on the VS. In case, VS is idle for a period of time as specified by ondemand_metrics_idle_timeout then metrics processing is suspended for that VS. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableOndemandMetrics *bool `json:"enable_ondemand_metrics,omitempty"`

	// Enable node (service engine) level analytics forvs metrics. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableSeAnalytics *bool `json:"enable_se_analytics,omitempty"`

	// Enables analytics on backend servers. This may be desired in container environment when there are large number of ephemeral servers. Additionally, no healthscore of servers is computed when server analytics is enabled. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableServerAnalytics *bool `json:"enable_server_analytics,omitempty"`

	// Enable VirtualService (frontend) Analytics. This flag enables metrics and healthscore for Virtualservice. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableVsAnalytics *bool `json:"enable_vs_analytics,omitempty"`

	// Exclude client closed connection before an HTTP request could be completed from being classified as an error. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeClientCloseBeforeRequestAsError *bool `json:"exclude_client_close_before_request_as_error,omitempty"`

	// Exclude dns policy drops from the list of errors. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeDNSPolicyDropAsSignificant *bool `json:"exclude_dns_policy_drop_as_significant,omitempty"`

	// Exclude queries to GSLB services that are operationally down from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeGsDownAsError *bool `json:"exclude_gs_down_as_error,omitempty"`

	// List of HTTP status codes to be excluded from being classified as an error.  Error connections or responses impacts health score, are included as significant logs, and may be classified as part of a DoS attack. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ExcludeHTTPErrorCodes []int64 `json:"exclude_http_error_codes,omitempty,omitempty"`

	// Exclude dns queries to domains outside the domains configured in the DNS application profile from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeInvalidDNSDomainAsError *bool `json:"exclude_invalid_dns_domain_as_error,omitempty"`

	// Exclude invalid dns queries from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeInvalidDNSQueryAsError *bool `json:"exclude_invalid_dns_query_as_error,omitempty"`

	// Exclude the Issuer-Revoked OCSP Responses from the list of errors. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	ExcludeIssuerRevokedOcspResponsesAsError *bool `json:"exclude_issuer_revoked_ocsp_responses_as_error,omitempty"`

	// Exclude queries to domains that did not have configured services/records from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeNoDNSRecordAsError *bool `json:"exclude_no_dns_record_as_error,omitempty"`

	// Exclude queries to GSLB services that have no available members from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeNoValidGsMemberAsError *bool `json:"exclude_no_valid_gs_member_as_error,omitempty"`

	// Exclude persistence server changed while load balancing' from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludePersistenceChangeAsError *bool `json:"exclude_persistence_change_as_error,omitempty"`

	// Exclude the Revoked OCSP certificate status responses from the list of errors. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	ExcludeRevokedOcspResponsesAsError *bool `json:"exclude_revoked_ocsp_responses_as_error,omitempty"`

	// Exclude server dns error response from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeServerDNSErrorAsError *bool `json:"exclude_server_dns_error_as_error,omitempty"`

	// Exclude server TCP reset from errors.  It is common for applications like MS Exchange. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeServerTCPResetAsError *bool `json:"exclude_server_tcp_reset_as_error,omitempty"`

	// List of SIP status codes to be excluded from being classified as an error. Field introduced in 17.2.13, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ExcludeSipErrorCodes []int64 `json:"exclude_sip_error_codes,omitempty,omitempty"`

	// Exclude the Stale OCSP certificate status responses from the list of errors. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	ExcludeStaleOcspResponsesAsError *bool `json:"exclude_stale_ocsp_responses_as_error,omitempty"`

	// Exclude 'server unanswered syns' from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeSynRetransmitAsError *bool `json:"exclude_syn_retransmit_as_error,omitempty"`

	// Exclude TCP resets by client from the list of potential errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeTCPResetAsError *bool `json:"exclude_tcp_reset_as_error,omitempty"`

	// Exclude the unavailable OCSP Responses from the list of errors. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	ExcludeUnavailableOcspResponsesAsError *bool `json:"exclude_unavailable_ocsp_responses_as_error,omitempty"`

	// Exclude unsupported dns queries from the list of errors. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ExcludeUnsupportedDNSQueryAsError *bool `json:"exclude_unsupported_dns_query_as_error,omitempty"`

	// Skips health score computation of pool servers when number of servers in a pool is more than this setting. Allowed values are 0-5000. Special values are 0- server health score is deactivated. Field introduced in 17.2.13, 18.1.4. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition. Special default for Essentials edition is 0, Basic edition is 0, Enterprise is 20.
	HealthscoreMaxServerLimit *uint32 `json:"healthscore_max_server_limit,omitempty"`

	// Time window (in secs) within which only unique health change events should occur. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1209600), Basic edition(Allowed values- 1209600), Enterprise with Cloud Services edition.
	HsEventThrottleWindow *uint32 `json:"hs_event_throttle_window,omitempty"`

	// Maximum penalty that may be deducted from health score for anomalies. Allowed values are 0-100. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 10), Basic edition(Allowed values- 10), Enterprise with Cloud Services edition.
	HsMaxAnomalyPenalty *uint32 `json:"hs_max_anomaly_penalty,omitempty"`

	// Maximum penalty that may be deducted from health score for high resource utilization. Allowed values are 0-100. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 25), Basic edition(Allowed values- 25), Enterprise with Cloud Services edition.
	HsMaxResourcesPenalty *uint32 `json:"hs_max_resources_penalty,omitempty"`

	// Maximum penalty that may be deducted from health score based on security assessment. Allowed values are 0-100. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 100), Basic edition(Allowed values- 100), Enterprise with Cloud Services edition.
	HsMaxSecurityPenalty *uint32 `json:"hs_max_security_penalty,omitempty"`

	// DoS connection rate below which the DoS security assessment will not kick in. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1000), Basic edition(Allowed values- 1000), Enterprise with Cloud Services edition.
	HsMinDosRate *uint32 `json:"hs_min_dos_rate,omitempty"`

	// Adds free performance score credits to health score. It can be used for compensating health score for known slow applications. Allowed values are 0-100. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	HsPerformanceBoost uint32 `json:"hs_performance_boost,omitempty"`

	// Threshold number of connections in 5min, below which apdexr, apdexc, rum_apdex, and other network quality metrics are not computed. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 10), Basic edition(Allowed values- 10), Enterprise with Cloud Services edition.
	HsPscoreTrafficThresholdL4Client *float64 `json:"hs_pscore_traffic_threshold_l4_client,omitempty"`

	// Threshold number of connections in 5min, below which apdexr, apdexc, rum_apdex, and other network quality metrics are not computed. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 10), Basic edition(Allowed values- 10), Enterprise with Cloud Services edition.
	HsPscoreTrafficThresholdL4Server *float64 `json:"hs_pscore_traffic_threshold_l4_server,omitempty"`

	// Score assigned when the certificate has expired. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0.0), Basic edition(Allowed values- 0.0), Enterprise with Cloud Services edition.
	HsSecurityCertscoreExpired *float64 `json:"hs_security_certscore_expired,omitempty"`

	// Score assigned when the certificate expires in more than 30 days. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5.0), Basic edition(Allowed values- 5.0), Enterprise with Cloud Services edition.
	HsSecurityCertscoreGt30d *float64 `json:"hs_security_certscore_gt30d,omitempty"`

	// Score assigned when the certificate expires in less than or equal to 7 days. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 2.0), Basic edition(Allowed values- 2.0), Enterprise with Cloud Services edition.
	HsSecurityCertscoreLe07d *float64 `json:"hs_security_certscore_le07d,omitempty"`

	// Score assigned when the certificate expires in less than or equal to 30 days. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 4.0), Basic edition(Allowed values- 4.0), Enterprise with Cloud Services edition.
	HsSecurityCertscoreLe30d *float64 `json:"hs_security_certscore_le30d,omitempty"`

	// Penalty for allowing certificates with invalid chain. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1.0), Basic edition(Allowed values- 1.0), Enterprise with Cloud Services edition.
	HsSecurityChainInvalidityPenalty *float64 `json:"hs_security_chain_invalidity_penalty,omitempty"`

	// Score assigned when the minimum cipher strength is 0 bits. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0.0), Basic edition(Allowed values- 0.0), Enterprise with Cloud Services edition.
	HsSecurityCipherscoreEq000b *float64 `json:"hs_security_cipherscore_eq000b,omitempty"`

	// Score assigned when the minimum cipher strength is greater than equal to 128 bits. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5.0), Basic edition(Allowed values- 5.0), Enterprise with Cloud Services edition.
	HsSecurityCipherscoreGe128b *float64 `json:"hs_security_cipherscore_ge128b,omitempty"`

	// Score assigned when the minimum cipher strength is less than 128 bits. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 3.5), Basic edition(Allowed values- 3.5), Enterprise with Cloud Services edition.
	HsSecurityCipherscoreLt128b *float64 `json:"hs_security_cipherscore_lt128b,omitempty"`

	// Score assigned when no algorithm is used for encryption. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0.0), Basic edition(Allowed values- 0.0), Enterprise with Cloud Services edition.
	HsSecurityEncalgoScoreNone *float64 `json:"hs_security_encalgo_score_none,omitempty"`

	// Score assigned when RC4 algorithm is used for encryption. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 2.5), Basic edition(Allowed values- 2.5), Enterprise with Cloud Services edition.
	HsSecurityEncalgoScoreRc4 *float64 `json:"hs_security_encalgo_score_rc4,omitempty"`

	// Penalty for not enabling HSTS. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1.0), Basic edition(Allowed values- 1.0), Enterprise with Cloud Services edition.
	HsSecurityHstsPenalty *float64 `json:"hs_security_hsts_penalty,omitempty"`

	// Penalty for allowing non-PFS handshakes. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1.0), Basic edition(Allowed values- 1.0), Enterprise with Cloud Services edition.
	HsSecurityNonpfsPenalty *float64 `json:"hs_security_nonpfs_penalty,omitempty"`

	// Score assigned when OCSP Certificate Status is set to Revoked or Issuer Revoked. Allowed values are 0.0-5.0. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0.0), Basic edition(Allowed values- 0.0), Enterprise with Cloud Services edition.
	HsSecurityOcspRevokedScore *float64 `json:"hs_security_ocsp_revoked_score,omitempty"`

	// Deprecated. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1.0), Basic edition(Allowed values- 1.0), Enterprise with Cloud Services edition.
	HsSecuritySelfsignedcertPenalty *float64 `json:"hs_security_selfsignedcert_penalty,omitempty"`

	// Score assigned when supporting SSL3.0 encryption protocol. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 3.5), Basic edition(Allowed values- 3.5), Enterprise with Cloud Services edition.
	HsSecuritySsl30Score *float64 `json:"hs_security_ssl30_score,omitempty"`

	// Score assigned when supporting TLS1.0 encryption protocol. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5.0), Basic edition(Allowed values- 5.0), Enterprise with Cloud Services edition.
	HsSecurityTLS10Score *float64 `json:"hs_security_tls10_score,omitempty"`

	// Score assigned when supporting TLS1.1 encryption protocol. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5.0), Basic edition(Allowed values- 5.0), Enterprise with Cloud Services edition.
	HsSecurityTLS11Score *float64 `json:"hs_security_tls11_score,omitempty"`

	// Score assigned when supporting TLS1.2 encryption protocol. Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5.0), Basic edition(Allowed values- 5.0), Enterprise with Cloud Services edition.
	HsSecurityTLS12Score *float64 `json:"hs_security_tls12_score,omitempty"`

	// Score assigned when supporting TLS1.3 encryption protocol. Allowed values are 0-5. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5.0), Basic edition(Allowed values- 5.0), Enterprise with Cloud Services edition.
	HsSecurityTLS13Score *float64 `json:"hs_security_tls13_score,omitempty"`

	// Penalty for allowing weak signature algorithm(s). Allowed values are 0-5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1.0), Basic edition(Allowed values- 1.0), Enterprise with Cloud Services edition.
	HsSecurityWeakSignatureAlgoPenalty *float64 `json:"hs_security_weak_signature_algo_penalty,omitempty"`

	// Deprecated in 22.1.1. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LatencyAuditProps *LatencyAuditProperties `json:"latency_audit_props,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// The name of the analytics profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// This flag sets the time duration of no live data traffic after which Virtual Service metrics processing is suspended. It is applicable only when enable_ondemand_metrics is set to false. Field introduced in 18.1.1. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OndemandMetricsIDLETimeout *uint32 `json:"ondemand_metrics_idle_timeout,omitempty"`

	// List of HTTP status code ranges to be excluded from being classified as an error. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ranges []*HttpstatusRange `json:"ranges,omitempty"`

	// Block of HTTP response codes to be excluded from being classified as an error. Enum options - AP_HTTP_RSP_4XX, AP_HTTP_RSP_5XX. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RespCodeBlock []string `json:"resp_code_block,omitempty"`

	// Rules applied to the HTTP application log for filtering sensitive information. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SensitiveLogProfile *SensitiveLogProfile `json:"sensitive_log_profile,omitempty"`

	// Maximum number of SIP messages added in logs for a SIP transaction. By default, this value is 20. Allowed values are 1-1000. Field introduced in 17.2.13, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 20), Basic edition(Allowed values- 20), Enterprise with Cloud Services edition.
	SipLogDepth *uint32 `json:"sip_log_depth,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Time Tracker Properties for connection establishment audit. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TimeTrackerProps *TimeTrackerProperties `json:"time_tracker_props,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the analytics profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

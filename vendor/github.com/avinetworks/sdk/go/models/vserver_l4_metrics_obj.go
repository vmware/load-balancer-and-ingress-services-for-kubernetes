package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VserverL4MetricsObj vserver l4 metrics obj
// swagger:model VserverL4MetricsObj
type VserverL4MetricsObj struct {

	// apdex measuring quality of network connections to servers.
	Apdexc *float64 `json:"apdexc,omitempty"`

	// apdex measuring network connection quality based on RTT.
	Apdexrtt *float64 `json:"apdexrtt,omitempty"`

	// Number of Application DDOS attacks occurring.
	AvgApplicationDosAttacks *float64 `json:"avg_application_dos_attacks,omitempty"`

	// Average transmit and receive network bandwidth between client and virtual service.
	AvgBandwidth *float64 `json:"avg_bandwidth,omitempty"`

	// Averaged rate bytes dropped per second.
	AvgBytesPolicyDrops *float64 `json:"avg_bytes_policy_drops,omitempty"`

	// Rate of total connections per second.
	AvgCompleteConns *float64 `json:"avg_complete_conns,omitempty"`

	// Rate of dropped connections per second.
	AvgConnectionsDropped *float64 `json:"avg_connections_dropped,omitempty"`

	// DoS attack  Rate of HTTP App Errors.
	AvgDosAppError *float64 `json:"avg_dos_app_error,omitempty"`

	// Number DDOS attacks occurring.
	AvgDosAttacks *float64 `json:"avg_dos_attacks,omitempty"`

	// DoS attack  Rate of Bad Rst Floods.
	AvgDosBadRstFlood *float64 `json:"avg_dos_bad_rst_flood,omitempty"`

	// Average transmit and receive network bandwidth between client and virtual service related to DDoS attack.
	AvgDosBandwidth *float64 `json:"avg_dos_bandwidth,omitempty"`

	// Number of connections considered as DoS.
	AvgDosConn *float64 `json:"avg_dos_conn,omitempty"`

	// DoS attack  Connections dropped due to IP rate limit.
	AvgDosConnIPRlDrop *float64 `json:"avg_dos_conn_ip_rl_drop,omitempty"`

	// DoS attack  Connections dropped due to VS rate limit.
	AvgDosConnRlDrop *float64 `json:"avg_dos_conn_rl_drop,omitempty"`

	// DoS attack  Rate of Fake Sessions.
	AvgDosFakeSession *float64 `json:"avg_dos_fake_session,omitempty"`

	// DoS attack  Rate of HTTP Aborts.
	AvgDosHTTPAbort *float64 `json:"avg_dos_http_abort,omitempty"`

	// DoS attack  Rate of HTTP Errors.
	AvgDosHTTPError *float64 `json:"avg_dos_http_error,omitempty"`

	// DoS attack  Rate of HTTP Timeouts.
	AvgDosHTTPTimeout *float64 `json:"avg_dos_http_timeout,omitempty"`

	// DoS attack  Rate of Malformed Packet Floods.
	AvgDosMalformedFlood *float64 `json:"avg_dos_malformed_flood,omitempty"`

	// DoS attack  Non SYN packet flood.
	AvgDosNonSynFlood *float64 `json:"avg_dos_non_syn_flood,omitempty"`

	// Number of request considered as DoS.
	AvgDosReq *float64 `json:"avg_dos_req,omitempty"`

	// DoS attack  Requests dropped due to Cookie rate limit.
	AvgDosReqCookieRlDrop *float64 `json:"avg_dos_req_cookie_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to Custom rate limit. Field introduced in 17.2.13,18.1.3,18.2.1.
	AvgDosReqCustomRlDrop *float64 `json:"avg_dos_req_custom_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to Header rate limit.
	AvgDosReqHdrRlDrop *float64 `json:"avg_dos_req_hdr_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP rate limit.
	AvgDosReqIPRlDrop *float64 `json:"avg_dos_req_ip_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP rate limit for bad requests.
	AvgDosReqIPRlDropBad *float64 `json:"avg_dos_req_ip_rl_drop_bad,omitempty"`

	// DoS attack  Requests dropped due to bad IP rate limit.
	AvgDosReqIPScanBadRlDrop *float64 `json:"avg_dos_req_ip_scan_bad_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to unknown IP rate limit.
	AvgDosReqIPScanUnknownRlDrop *float64 `json:"avg_dos_req_ip_scan_unknown_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP+URL rate limit.
	AvgDosReqIPURIRlDrop *float64 `json:"avg_dos_req_ip_uri_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP+URL rate limit for bad requests.
	AvgDosReqIPURIRlDropBad *float64 `json:"avg_dos_req_ip_uri_rl_drop_bad,omitempty"`

	// DoS attack  Requests dropped due to VS rate limit.
	AvgDosReqRlDrop *float64 `json:"avg_dos_req_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to URL rate limit.
	AvgDosReqURIRlDrop *float64 `json:"avg_dos_req_uri_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to URL rate limit for bad requests.
	AvgDosReqURIRlDropBad *float64 `json:"avg_dos_req_uri_rl_drop_bad,omitempty"`

	// DoS attack  Requests dropped due to bad URL rate limit.
	AvgDosReqURIScanBadRlDrop *float64 `json:"avg_dos_req_uri_scan_bad_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to unknown URL rate limit.
	AvgDosReqURIScanUnknownRlDrop *float64 `json:"avg_dos_req_uri_scan_unknown_rl_drop,omitempty"`

	// Average rate of bytes received per second related to DDoS attack.
	AvgDosRxBytes *float64 `json:"avg_dos_rx_bytes,omitempty"`

	// DoS attack  Slow Uri.
	AvgDosSlowURI *float64 `json:"avg_dos_slow_uri,omitempty"`

	// DoS attack  Rate of Small Window Stresses.
	AvgDosSmallWindowStress *float64 `json:"avg_dos_small_window_stress,omitempty"`

	// DoS attack  Rate of HTTP SSL Errors.
	AvgDosSslError *float64 `json:"avg_dos_ssl_error,omitempty"`

	// DoS attack  Rate of Syn Floods.
	AvgDosSynFlood *float64 `json:"avg_dos_syn_flood,omitempty"`

	// Total number of request used for L7 dos requests normalization.
	AvgDosTotalReq *float64 `json:"avg_dos_total_req,omitempty"`

	// Average rate of bytes transmitted per second related to DDoS attack.
	AvgDosTxBytes *float64 `json:"avg_dos_tx_bytes,omitempty"`

	// DoS attack  Rate of Zero Window Stresses.
	AvgDosZeroWindowStress *float64 `json:"avg_dos_zero_window_stress,omitempty"`

	// Rate of total errored connections per second.
	AvgErroredConnections *float64 `json:"avg_errored_connections,omitempty"`

	// Average L4 connection duration which does not include client RTT.
	AvgL4ClientLatency *float64 `json:"avg_l4_client_latency,omitempty"`

	// Rate of lossy connections per second.
	AvgLossyConnections *float64 `json:"avg_lossy_connections,omitempty"`

	// Averaged rate of lossy request per second.
	AvgLossyReq *float64 `json:"avg_lossy_req,omitempty"`

	// Number of Network DDOS attacks occurring.
	AvgNetworkDosAttacks *float64 `json:"avg_network_dos_attacks,omitempty"`

	// Averaged rate of new client connections per second.
	AvgNewEstablishedConns *float64 `json:"avg_new_established_conns,omitempty"`

	// Averaged rate of dropped packets per second due to policy.
	AvgPktsPolicyDrops *float64 `json:"avg_pkts_policy_drops,omitempty"`

	// Rate of total connections dropped due to VS policy per second. It includes drops due to rate limits, security policy drops, connection limits etc.
	AvgPolicyDrops *float64 `json:"avg_policy_drops,omitempty"`

	// Average rate of bytes received per second.
	AvgRxBytes *float64 `json:"avg_rx_bytes,omitempty"`

	// Average rate of received bytes dropped per second.
	AvgRxBytesDropped *float64 `json:"avg_rx_bytes_dropped,omitempty"`

	// Average rate of packets received per second.
	AvgRxPkts *float64 `json:"avg_rx_pkts,omitempty"`

	// Average rate of received packets dropped per second.
	AvgRxPktsDropped *float64 `json:"avg_rx_pkts_dropped,omitempty"`

	// Total syncs sent across all connections.
	AvgSyns *float64 `json:"avg_syns,omitempty"`

	// Averaged rate bytes dropped per second.
	AvgTotalConnections *float64 `json:"avg_total_connections,omitempty"`

	// Average network round trip time between client and virtual service.
	AvgTotalRtt *float64 `json:"avg_total_rtt,omitempty"`

	// Average rate of bytes transmitted per second.
	AvgTxBytes *float64 `json:"avg_tx_bytes,omitempty"`

	// Average rate of packets transmitted per second.
	AvgTxPkts *float64 `json:"avg_tx_pkts,omitempty"`

	// Max number of SEs.
	MaxNumActiveSe *float64 `json:"max_num_active_se,omitempty"`

	// Max number of open connections.
	MaxOpenConns *float64 `json:"max_open_conns,omitempty"`

	// Total number of received bytes.
	MaxRxBytesAbsolute *float64 `json:"max_rx_bytes_absolute,omitempty"`

	// Total number of received frames.
	MaxRxPktsAbsolute *float64 `json:"max_rx_pkts_absolute,omitempty"`

	// Total number of transmitted bytes.
	MaxTxBytesAbsolute *float64 `json:"max_tx_bytes_absolute,omitempty"`

	// Total number of transmitted frames.
	MaxTxPktsAbsolute *float64 `json:"max_tx_pkts_absolute,omitempty"`

	// node_obj_id of VserverL4MetricsObj.
	// Required: true
	NodeObjID *string `json:"node_obj_id"`

	// Fraction of L7 requests owing to DoS.
	PctApplicationDosAttacks *float64 `json:"pct_application_dos_attacks,omitempty"`

	// Percent of l4 connection dropped and lossy for virtual service.
	PctConnectionErrors *float64 `json:"pct_connection_errors,omitempty"`

	// Fraction of L4 connections owing to DoS.
	PctConnectionsDosAttacks *float64 `json:"pct_connections_dos_attacks,omitempty"`

	// DoS bandwidth percentage.
	PctDosBandwidth *float64 `json:"pct_dos_bandwidth,omitempty"`

	// Percentage of received bytes as part of a DoS attack.
	PctDosRxBytes *float64 `json:"pct_dos_rx_bytes,omitempty"`

	// Deprecated.
	PctNetworkDosAttacks *float64 `json:"pct_network_dos_attacks,omitempty"`

	// Fraction of packets owing to DoS.
	PctPktsDosAttacks *float64 `json:"pct_pkts_dos_attacks,omitempty"`

	// Fraction of L4 requests dropped owing to policy.
	PctPolicyDrops *float64 `json:"pct_policy_drops,omitempty"`

	// Total duration across all connections.
	SumConnDuration *float64 `json:"sum_conn_duration,omitempty"`

	// Total number of connection dropped due to vserver connection limit.
	SumConnectionDroppedUserLimit *float64 `json:"sum_connection_dropped_user_limit,omitempty"`

	// Total number of client network connections that were lossy or dropped.
	SumConnectionErrors *float64 `json:"sum_connection_errors,omitempty"`

	// Total connections dropped including failed.
	SumConnectionsDropped *float64 `json:"sum_connections_dropped,omitempty"`

	// Total duplicate ACK retransmits across all connections.
	SumDupAckRetransmits *float64 `json:"sum_dup_ack_retransmits,omitempty"`

	// Sum of end to end network RTT experienced by end clients. Higher value would increase response times experienced by clients.
	SumEndToEndRtt *float64 `json:"sum_end_to_end_rtt,omitempty"`

	// Total connections that have RTT values from 0 to RTT threshold.
	SumEndToEndRttBucket1 *float64 `json:"sum_end_to_end_rtt_bucket1,omitempty"`

	// Total connections that have RTT values RTT threshold and above.
	SumEndToEndRttBucket2 *float64 `json:"sum_end_to_end_rtt_bucket2,omitempty"`

	// Total number of finished connections.
	SumFinishedConns *float64 `json:"sum_finished_conns,omitempty"`

	// Total connections that were lossy due to high packet retransmissions.
	SumLossyConnections *float64 `json:"sum_lossy_connections,omitempty"`

	// Total requests that were lossy due to high packet retransmissions.
	SumLossyReq *float64 `json:"sum_lossy_req,omitempty"`

	// Total out of order packets across all connections.
	SumOutOfOrders *float64 `json:"sum_out_of_orders,omitempty"`

	// Total number of packets dropped due to vserver bandwidth limit.
	SumPacketDroppedUserBandwidthLimit *float64 `json:"sum_packet_dropped_user_bandwidth_limit,omitempty"`

	// Total number connections used for rtt.
	SumRttValidConnections *float64 `json:"sum_rtt_valid_connections,omitempty"`

	// Total SACK retransmits across all connections.
	SumSackRetransmits *float64 `json:"sum_sack_retransmits,omitempty"`

	// Total number of connections with server flow control condition.
	SumServerFlowControl *float64 `json:"sum_server_flow_control,omitempty"`

	// Total connection timeouts in the interval.
	SumTimeoutRetransmits *float64 `json:"sum_timeout_retransmits,omitempty"`

	// Total number of zero window size events across all connections.
	SumZeroWindowSizeEvents *float64 `json:"sum_zero_window_size_events,omitempty"`
}

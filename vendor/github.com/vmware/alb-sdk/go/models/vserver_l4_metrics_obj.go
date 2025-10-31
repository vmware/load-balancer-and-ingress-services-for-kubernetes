// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VserverL4MetricsObj vserver l4 metrics obj
// swagger:model VserverL4MetricsObj
type VserverL4MetricsObj struct {

	// apdex measuring quality of network connections to servers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Apdexc *float64 `json:"apdexc,omitempty"`

	// apdex measuring network connection quality based on RTT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Apdexrtt *float64 `json:"apdexrtt,omitempty"`

	// Number of Application DDOS attacks occurring. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgApplicationDosAttacks *float64 `json:"avg_application_dos_attacks,omitempty"`

	// Average transmit and receive network bandwidth between client and virtual service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgBandwidth *float64 `json:"avg_bandwidth,omitempty"`

	// Averaged rate bytes dropped per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgBytesPolicyDrops *float64 `json:"avg_bytes_policy_drops,omitempty"`

	// Rate of total connections per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgCompleteConns *float64 `json:"avg_complete_conns,omitempty"`

	// Rate of dropped connections per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgConnectionsDropped *float64 `json:"avg_connections_dropped,omitempty"`

	// DoS attack  Rate of HTTP App Errors. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosAppError *float64 `json:"avg_dos_app_error,omitempty"`

	// Number DDOS attacks occurring. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosAttacks *float64 `json:"avg_dos_attacks,omitempty"`

	// DoS attack  Rate of Bad Rst Floods. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosBadRstFlood *float64 `json:"avg_dos_bad_rst_flood,omitempty"`

	// Average transmit and receive network bandwidth between client and virtual service related to DDoS attack. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosBandwidth *float64 `json:"avg_dos_bandwidth,omitempty"`

	// Number of connections considered as DoS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosConn *float64 `json:"avg_dos_conn,omitempty"`

	// DoS attack  Connections dropped due to IP rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosConnIPRlDrop *float64 `json:"avg_dos_conn_ip_rl_drop,omitempty"`

	// DoS attack  Connections dropped due to VS rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosConnRlDrop *float64 `json:"avg_dos_conn_rl_drop,omitempty"`

	// DoS attack  Rate of Fake Sessions. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosFakeSession *float64 `json:"avg_dos_fake_session,omitempty"`

	// DoS attack  Rate of HTTP Aborts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosHTTPAbort *float64 `json:"avg_dos_http_abort,omitempty"`

	// DoS attack  Rate of HTTP Errors. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosHTTPError *float64 `json:"avg_dos_http_error,omitempty"`

	// DoS attack  Rate of HTTP Timeouts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosHTTPTimeout *float64 `json:"avg_dos_http_timeout,omitempty"`

	// DoS attack  Rate of Malformed Packet Floods. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosMalformedFlood *float64 `json:"avg_dos_malformed_flood,omitempty"`

	// DoS attack  Non SYN packet flood. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosNonSynFlood *float64 `json:"avg_dos_non_syn_flood,omitempty"`

	// Number of request considered as DoS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReq *float64 `json:"avg_dos_req,omitempty"`

	// DoS attack  Requests dropped due to Cookie rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqCookieRlDrop *float64 `json:"avg_dos_req_cookie_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to Custom rate limit. Field introduced in 17.2.13,18.1.3,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqCustomRlDrop *float64 `json:"avg_dos_req_custom_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to Header rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqHdrRlDrop *float64 `json:"avg_dos_req_hdr_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqIPRlDrop *float64 `json:"avg_dos_req_ip_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP rate limit for bad requests. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqIPRlDropBad *float64 `json:"avg_dos_req_ip_rl_drop_bad,omitempty"`

	// DoS attack  Requests dropped due to bad IP rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqIPScanBadRlDrop *float64 `json:"avg_dos_req_ip_scan_bad_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to unknown IP rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqIPScanUnknownRlDrop *float64 `json:"avg_dos_req_ip_scan_unknown_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP+URL rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqIPURIRlDrop *float64 `json:"avg_dos_req_ip_uri_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to IP+URL rate limit for bad requests. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqIPURIRlDropBad *float64 `json:"avg_dos_req_ip_uri_rl_drop_bad,omitempty"`

	// DoS attack  Requests dropped due to VS rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqRlDrop *float64 `json:"avg_dos_req_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to URL rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqURIRlDrop *float64 `json:"avg_dos_req_uri_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to URL rate limit for bad requests. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqURIRlDropBad *float64 `json:"avg_dos_req_uri_rl_drop_bad,omitempty"`

	// DoS attack  Requests dropped due to bad URL rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqURIScanBadRlDrop *float64 `json:"avg_dos_req_uri_scan_bad_rl_drop,omitempty"`

	// DoS attack  Requests dropped due to unknown URL rate limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosReqURIScanUnknownRlDrop *float64 `json:"avg_dos_req_uri_scan_unknown_rl_drop,omitempty"`

	// Average rate of bytes received per second related to DDoS attack. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosRxBytes *float64 `json:"avg_dos_rx_bytes,omitempty"`

	// DoS attack  Slow Uri. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosSlowURI *float64 `json:"avg_dos_slow_uri,omitempty"`

	// DoS attack  Rate of Small Window Stresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosSmallWindowStress *float64 `json:"avg_dos_small_window_stress,omitempty"`

	// DoS attack  Rate of HTTP SSL Errors. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosSslError *float64 `json:"avg_dos_ssl_error,omitempty"`

	// DoS attack  Rate of Syn Floods. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosSynFlood *float64 `json:"avg_dos_syn_flood,omitempty"`

	// Total number of request used for L7 dos requests normalization. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosTotalReq *float64 `json:"avg_dos_total_req,omitempty"`

	// Average rate of bytes transmitted per second related to DDoS attack. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosTxBytes *float64 `json:"avg_dos_tx_bytes,omitempty"`

	// DoS attack  Rate of Zero Window Stresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgDosZeroWindowStress *float64 `json:"avg_dos_zero_window_stress,omitempty"`

	// Rate of total errored connections per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgErroredConnections *float64 `json:"avg_errored_connections,omitempty"`

	// Average rate of SYN DDoS attacks on Virtual Service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgHalfOpenConns *float64 `json:"avg_half_open_conns,omitempty"`

	// Average L4 connection duration which does not include client RTT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgL4ClientLatency *float64 `json:"avg_l4_client_latency,omitempty"`

	// Rate of lossy connections per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgLossyConnections *float64 `json:"avg_lossy_connections,omitempty"`

	// Averaged rate of lossy request per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgLossyReq *float64 `json:"avg_lossy_req,omitempty"`

	// Number of Network DDOS attacks occurring. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgNetworkDosAttacks *float64 `json:"avg_network_dos_attacks,omitempty"`

	// Averaged rate of new client connections per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgNewEstablishedConns *float64 `json:"avg_new_established_conns,omitempty"`

	// Averaged rate of dropped packets per second due to policy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgPktsPolicyDrops *float64 `json:"avg_pkts_policy_drops,omitempty"`

	// Rate of total connections dropped due to VS policy per second. It includes drops due to rate limits, security policy drops, connection limits etc. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgPolicyDrops *float64 `json:"avg_policy_drops,omitempty"`

	// Average rate of bytes received per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgRxBytes *float64 `json:"avg_rx_bytes,omitempty"`

	// Average rate of received bytes dropped per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgRxBytesDropped *float64 `json:"avg_rx_bytes_dropped,omitempty"`

	// Average rate of packets received per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgRxPkts *float64 `json:"avg_rx_pkts,omitempty"`

	// Average rate of received packets dropped per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgRxPktsDropped *float64 `json:"avg_rx_pkts_dropped,omitempty"`

	// Total syncs sent across all connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgSyns *float64 `json:"avg_syns,omitempty"`

	// Averaged rate bytes dropped per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgTotalConnections *float64 `json:"avg_total_connections,omitempty"`

	// Average network round trip time between client and virtual service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgTotalRtt *float64 `json:"avg_total_rtt,omitempty"`

	// Average rate of bytes transmitted per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgTxBytes *float64 `json:"avg_tx_bytes,omitempty"`

	// Average rate of packets transmitted per second. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgTxPkts *float64 `json:"avg_tx_pkts,omitempty"`

	// Maximum connection establishment time on the client side. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxConnectionEstbTimeFe *float64 `json:"max_connection_estb_time_fe,omitempty"`

	// Max number of SEs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxNumActiveSe *float64 `json:"max_num_active_se,omitempty"`

	// Max number of open connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxOpenConns *float64 `json:"max_open_conns,omitempty"`

	// Total number of received bytes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRxBytesAbsolute *float64 `json:"max_rx_bytes_absolute,omitempty"`

	// Total number of received frames. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRxPktsAbsolute *float64 `json:"max_rx_pkts_absolute,omitempty"`

	// Total number of transmitted bytes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxTxBytesAbsolute *float64 `json:"max_tx_bytes_absolute,omitempty"`

	// Total number of transmitted frames. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxTxPktsAbsolute *float64 `json:"max_tx_pkts_absolute,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NodeObjID *string `json:"node_obj_id"`

	// Fraction of L7 requests owing to DoS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctApplicationDosAttacks *float64 `json:"pct_application_dos_attacks,omitempty"`

	// Percent of l4 connection dropped and lossy for virtual service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctConnectionErrors *float64 `json:"pct_connection_errors,omitempty"`

	// Fraction of L4 connections owing to DoS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctConnectionsDosAttacks *float64 `json:"pct_connections_dos_attacks,omitempty"`

	// DoS bandwidth percentage. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctDosBandwidth *float64 `json:"pct_dos_bandwidth,omitempty"`

	// Percentage of received bytes as part of a DoS attack. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctDosRxBytes *float64 `json:"pct_dos_rx_bytes,omitempty"`

	// Deprecated. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctNetworkDosAttacks *float64 `json:"pct_network_dos_attacks,omitempty"`

	// Fraction of packets owing to DoS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctPktsDosAttacks *float64 `json:"pct_pkts_dos_attacks,omitempty"`

	// Fraction of L4 requests dropped owing to policy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PctPolicyDrops *float64 `json:"pct_policy_drops,omitempty"`

	// Total duration across all connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumConnDuration *float64 `json:"sum_conn_duration,omitempty"`

	// Total number of times client side connection establishment time was breached. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SumConnEstTimeExceededFlowsFe *float64 `json:"sum_conn_est_time_exceeded_flows_fe,omitempty"`

	// Total number of connection dropped due to vserver connection limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumConnectionDroppedUserLimit *float64 `json:"sum_connection_dropped_user_limit,omitempty"`

	// Total number of client network connections that were lossy or dropped. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumConnectionErrors *float64 `json:"sum_connection_errors,omitempty"`

	// Total connections dropped including failed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumConnectionsDropped *float64 `json:"sum_connections_dropped,omitempty"`

	// Total duplicate ACK retransmits across all connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumDupAckRetransmits *float64 `json:"sum_dup_ack_retransmits,omitempty"`

	// Sum of end to end network RTT experienced by end clients. Higher value would increase response times experienced by clients. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumEndToEndRtt *float64 `json:"sum_end_to_end_rtt,omitempty"`

	// Total connections that have RTT values from 0 to RTT threshold. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumEndToEndRttBucket1 *float64 `json:"sum_end_to_end_rtt_bucket1,omitempty"`

	// Total connections that have RTT values RTT threshold and above. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumEndToEndRttBucket2 *float64 `json:"sum_end_to_end_rtt_bucket2,omitempty"`

	// Total number of finished connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumFinishedConns *float64 `json:"sum_finished_conns,omitempty"`

	// Total number of times 'latency_threshold' was breached during ingress. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SumIngressLatencyExceededFlows *float64 `json:"sum_ingress_latency_exceeded_flows,omitempty"`

	// Total connections that were lossy due to high packet retransmissions. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumLossyConnections *float64 `json:"sum_lossy_connections,omitempty"`

	// Total requests that were lossy due to high packet retransmissions. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumLossyReq *float64 `json:"sum_lossy_req,omitempty"`

	// Total out of order packets across all connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumOutOfOrders *float64 `json:"sum_out_of_orders,omitempty"`

	// Total number of packets dropped due to vserver bandwidth limit. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumPacketDroppedUserBandwidthLimit *float64 `json:"sum_packet_dropped_user_bandwidth_limit,omitempty"`

	// Total number connections used for rtt. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumRttValidConnections *float64 `json:"sum_rtt_valid_connections,omitempty"`

	// Total SACK retransmits across all connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumSackRetransmits *float64 `json:"sum_sack_retransmits,omitempty"`

	// Total number of connections with server flow control condition. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumServerFlowControl *float64 `json:"sum_server_flow_control,omitempty"`

	// Total connection timeouts in the interval. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumTimeoutRetransmits *float64 `json:"sum_timeout_retransmits,omitempty"`

	// Total number of zero window size events across all connections. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumZeroWindowSizeEvents *float64 `json:"sum_zero_window_size_events,omitempty"`
}

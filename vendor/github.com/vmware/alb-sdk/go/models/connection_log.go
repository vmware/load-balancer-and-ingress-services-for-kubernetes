// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConnectionLog connection log
// swagger:model ConnectionLog
type ConnectionLog struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Adf *bool `json:"adf"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AverageTurntime *int32 `json:"average_turntime,omitempty"`

	// Average packet processing latency for the backend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AvgIngressLatencyBe *int32 `json:"avg_ingress_latency_be,omitempty"`

	// Average packet processing latency for the frontend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AvgIngressLatencyFe *int32 `json:"avg_ingress_latency_fe,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientDestPort *int32 `json:"client_dest_port"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientIP *int32 `json:"client_ip"`

	// IPv6 address of the client. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientIp6 *string `json:"client_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientLocation *string `json:"client_location,omitempty"`

	// Name of the Client Log Filter applied. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientLogFilterName *string `json:"client_log_filter_name,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientRtt *int32 `json:"client_rtt"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ClientSrcPort *int32 `json:"client_src_port"`

	// TCP connection establishment time for the backend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ConnEstTimeBe *int32 `json:"conn_est_time_be,omitempty"`

	// TCP connection establishment time for the frontend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ConnEstTimeFe *int32 `json:"conn_est_time_fe,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ConnectionEnded *bool `json:"connection_ended"`

	//  Enum options - DNS_ENTRY_PASS_THROUGH, DNS_ENTRY_GSLB, DNS_ENTRY_VIRTUALSERVICE, DNS_ENTRY_STATIC, DNS_ENTRY_POLICY, DNS_ENTRY_LOCAL. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSEtype *string `json:"dns_etype,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSFqdn *string `json:"dns_fqdn,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSIps []int64 `json:"dns_ips,omitempty,omitempty"`

	//  Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSQtype *string `json:"dns_qtype,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSRequest *DNSRequest `json:"dns_request,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSResponse *DNSResponse `json:"dns_response,omitempty"`

	// Datascript Log. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DsLog *string `json:"ds_log,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbpoolName *string `json:"gslbpool_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Gslbservice *string `json:"gslbservice,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbserviceName *string `json:"gslbservice_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	LogID *int32 `json:"log_id"`

	// Maximum packet processing latency for the backend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxIngressLatencyBe *int32 `json:"max_ingress_latency_be,omitempty"`

	// Maximum packet processing latency for the frontend flow. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxIngressLatencyFe *int32 `json:"max_ingress_latency_fe,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Microservice *string `json:"microservice,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MicroserviceName *string `json:"microservice_name,omitempty"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Mss *int32 `json:"mss"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkSecurityPolicyRuleName *string `json:"network_security_policy_rule_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSynRetransmit *int32 `json:"num_syn_retransmit,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumTransaction *int32 `json:"num_transaction,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumWindowShrink *int32 `json:"num_window_shrink,omitempty"`

	// OCSP Response sent in the SSL/TLS connection Handshake. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OcspStatusRespSent *bool `json:"ocsp_status_resp_sent,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	OutOfOrders *int32 `json:"out_of_orders"`

	// Persistence applied during server selection. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PersistenceUsed *bool `json:"persistence_used,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Pool *string `json:"pool,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	//  Enum options - PROTOCOL_ICMP, PROTOCOL_TCP, PROTOCOL_UDP, PROTOCOL_SCTP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Protocol *string `json:"protocol,omitempty"`

	// Version of proxy protocol used to convey client connection information to the back-end servers.  A value of 0 indicates that proxy protocol is not used.  A value of 1 or 2 indicates the version of proxy protocol used. Enum options - PROXY_PROTOCOL_VERSION_1, PROXY_PROTOCOL_VERSION_2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProxyProtocol *string `json:"proxy_protocol,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ReportTimestamp *int64 `json:"report_timestamp"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Retransmits *int32 `json:"retransmits"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RxBytes *int64 `json:"rx_bytes"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RxPkts *int64 `json:"rx_pkts"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerConnSrcIP *int32 `json:"server_conn_src_ip"`

	// IPv6 address used to connect to Backend Server. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerConnSrcIp6 *string `json:"server_conn_src_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerDestPort *int32 `json:"server_dest_port"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerIP *int32 `json:"server_ip"`

	// IPv6 address of the Backend Server. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerIp6 *string `json:"server_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerName *string `json:"server_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerNumWindowShrink *int32 `json:"server_num_window_shrink,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerOutOfOrders *int32 `json:"server_out_of_orders"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerRetransmits *int32 `json:"server_retransmits"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerRtt *int32 `json:"server_rtt"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerRxBytes *int64 `json:"server_rx_bytes"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerRxPkts *int64 `json:"server_rx_pkts"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerSrcPort *int32 `json:"server_src_port"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerTimeouts *int32 `json:"server_timeouts"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerTotalBytes *int64 `json:"server_total_bytes"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerTotalPkts *int64 `json:"server_total_pkts"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerTxBytes *int64 `json:"server_tx_bytes"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerTxPkts *int64 `json:"server_tx_pkts"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerZeroWindowSizeEvents *int32 `json:"server_zero_window_size_events"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceEngine *string `json:"service_engine,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Significance *string `json:"significance,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Significant *int64 `json:"significant"`

	// List of enums which indicate why a log is significant. Enum options - ADF_CLIENT_CONN_SETUP_REFUSED, ADF_SERVER_CONN_SETUP_REFUSED, ADF_CLIENT_CONN_SETUP_TIMEDOUT, ADF_SERVER_CONN_SETUP_TIMEDOUT, ADF_CLIENT_CONN_SETUP_FAILED_INTERNAL, ADF_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_CONN_SETUP_FAILED_BAD_PACKET, ADF_UDP_CONN_SETUP_FAILED_INTERNAL, ADF_UDP_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_SENT_RESET, ADF_SERVER_SENT_RESET, ADF_CLIENT_CONN_TIMEDOUT, ADF_SERVER_CONN_TIMEDOUT, ADF_USER_DELETE_OPERATION, ADF_CLIENT_REQUEST_TIMEOUT, ADF_CLIENT_CONN_ABORTED, ADF_CLIENT_SSL_HANDSHAKE_FAILURE, ADF_CLIENT_CONN_FAILED, ADF_SERVER_CERTIFICATE_VERIFICATION_FAILED, ADF_SERVER_SIDE_SSL_HANDSHAKE_FAILED.... Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SignificantLog []string `json:"significant_log,omitempty"`

	// SIP related logging information. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SipLog *SipLog `json:"sip_log,omitempty"`

	//  Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SniHostname *string `json:"sni_hostname,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslCipher *string `json:"ssl_cipher,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslSessionID *string `json:"ssl_session_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslVersion *string `json:"ssl_version,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	StartTimestamp *int64 `json:"start_timestamp"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Timeouts *int32 `json:"timeouts"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TotalBytes *int64 `json:"total_bytes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TotalPkts *int64 `json:"total_pkts,omitempty"`

	//  Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TotalTime *int64 `json:"total_time,omitempty"`

	//  Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	TxBytes *int64 `json:"tx_bytes"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	TxPkts *int64 `json:"tx_pkts"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Udf *bool `json:"udf"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VcpuID *int32 `json:"vcpu_id"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Virtualservice *string `json:"virtualservice"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsIP *int32 `json:"vs_ip,omitempty"`

	// IPv6 address of the VIP of the VS. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsIp6 *string `json:"vs_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ZeroWindowSizeEvents *int32 `json:"zero_window_size_events"`
}

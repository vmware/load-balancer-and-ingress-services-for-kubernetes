package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConnectionLog connection log
// swagger:model ConnectionLog
type ConnectionLog struct {

	// Placeholder for description of property adf of obj type ConnectionLog field type str  type boolean
	// Required: true
	Adf *bool `json:"adf"`

	//  Unit is MILLISECONDS.
	AverageTurntime *int32 `json:"average_turntime,omitempty"`

	// Number of client_dest_port.
	// Required: true
	ClientDestPort *int32 `json:"client_dest_port"`

	// Number of client_ip.
	// Required: true
	ClientIP *int32 `json:"client_ip"`

	// IPv6 address of the client. Field introduced in 18.1.1.
	ClientIp6 *string `json:"client_ip6,omitempty"`

	// client_location of ConnectionLog.
	ClientLocation *string `json:"client_location,omitempty"`

	// Name of the Client Log Filter applied. Field introduced in 18.1.5, 18.2.1.
	ClientLogFilterName *string `json:"client_log_filter_name,omitempty"`

	//  Unit is MILLISECONDS.
	// Required: true
	ClientRtt *int32 `json:"client_rtt"`

	// Number of client_src_port.
	// Required: true
	ClientSrcPort *int32 `json:"client_src_port"`

	// Placeholder for description of property connection_ended of obj type ConnectionLog field type str  type boolean
	// Required: true
	ConnectionEnded *bool `json:"connection_ended"`

	//  Enum options - DNS_ENTRY_PASS_THROUGH, DNS_ENTRY_GSLB, DNS_ENTRY_VIRTUALSERVICE, DNS_ENTRY_STATIC, DNS_ENTRY_POLICY, DNS_ENTRY_LOCAL.
	DNSEtype *string `json:"dns_etype,omitempty"`

	// dns_fqdn of ConnectionLog.
	DNSFqdn *string `json:"dns_fqdn,omitempty"`

	// Number of dns_ips.
	DNSIps []int64 `json:"dns_ips,omitempty,omitempty"`

	//  Enum options - DNS_RECORD_OTHER, DNS_RECORD_A, DNS_RECORD_NS, DNS_RECORD_CNAME, DNS_RECORD_SOA, DNS_RECORD_PTR, DNS_RECORD_HINFO, DNS_RECORD_MX, DNS_RECORD_TXT, DNS_RECORD_RP, DNS_RECORD_DNSKEY, DNS_RECORD_AAAA, DNS_RECORD_SRV, DNS_RECORD_OPT, DNS_RECORD_RRSIG, DNS_RECORD_AXFR, DNS_RECORD_ANY.
	DNSQtype *string `json:"dns_qtype,omitempty"`

	//  Field introduced in 17.1.1.
	DNSRequest *DNSRequest `json:"dns_request,omitempty"`

	// Placeholder for description of property dns_response of obj type ConnectionLog field type str  type object
	DNSResponse *DNSResponse `json:"dns_response,omitempty"`

	// Datascript Log. Field introduced in 18.2.3.
	DsLog *string `json:"ds_log,omitempty"`

	// gslbpool_name of ConnectionLog.
	GslbpoolName *string `json:"gslbpool_name,omitempty"`

	// gslbservice of ConnectionLog.
	Gslbservice *string `json:"gslbservice,omitempty"`

	// gslbservice_name of ConnectionLog.
	GslbserviceName *string `json:"gslbservice_name,omitempty"`

	// Number of log_id.
	// Required: true
	LogID *int32 `json:"log_id"`

	// microservice of ConnectionLog.
	Microservice *string `json:"microservice,omitempty"`

	// microservice_name of ConnectionLog.
	MicroserviceName *string `json:"microservice_name,omitempty"`

	//  Unit is BYTES.
	// Required: true
	Mss *int32 `json:"mss"`

	// network_security_policy_rule_name of ConnectionLog.
	NetworkSecurityPolicyRuleName *string `json:"network_security_policy_rule_name,omitempty"`

	// Number of num_syn_retransmit.
	NumSynRetransmit *int32 `json:"num_syn_retransmit,omitempty"`

	// Number of num_transaction.
	NumTransaction *int32 `json:"num_transaction,omitempty"`

	// Number of num_window_shrink.
	NumWindowShrink *int32 `json:"num_window_shrink,omitempty"`

	// OCSP Response sent in the SSL/TLS connection Handshake. Field introduced in 20.1.1.
	OcspStatusRespSent *bool `json:"ocsp_status_resp_sent,omitempty"`

	// Number of out_of_orders.
	// Required: true
	OutOfOrders *int32 `json:"out_of_orders"`

	// Persistence applied during server selection. Field introduced in 20.1.1.
	PersistenceUsed *bool `json:"persistence_used,omitempty"`

	// pool of ConnectionLog.
	Pool *string `json:"pool,omitempty"`

	// pool_name of ConnectionLog.
	PoolName *string `json:"pool_name,omitempty"`

	//  Enum options - PROTOCOL_ICMP, PROTOCOL_TCP, PROTOCOL_UDP.
	Protocol *string `json:"protocol,omitempty"`

	// Version of proxy protocol used to convey client connection information to the back-end servers.  A value of 0 indicates that proxy protocol is not used.  A value of 1 or 2 indicates the version of proxy protocol used. Enum options - PROXY_PROTOCOL_VERSION_1, PROXY_PROTOCOL_VERSION_2.
	ProxyProtocol *string `json:"proxy_protocol,omitempty"`

	// Number of report_timestamp.
	// Required: true
	ReportTimestamp *int64 `json:"report_timestamp"`

	// Number of retransmits.
	// Required: true
	Retransmits *int32 `json:"retransmits"`

	//  Unit is BYTES.
	// Required: true
	RxBytes *int64 `json:"rx_bytes"`

	// Number of rx_pkts.
	// Required: true
	RxPkts *int64 `json:"rx_pkts"`

	// Number of server_conn_src_ip.
	// Required: true
	ServerConnSrcIP *int32 `json:"server_conn_src_ip"`

	// IPv6 address used to connect to Backend Server. Field introduced in 18.1.1.
	ServerConnSrcIp6 *string `json:"server_conn_src_ip6,omitempty"`

	// Number of server_dest_port.
	// Required: true
	ServerDestPort *int32 `json:"server_dest_port"`

	// Number of server_ip.
	// Required: true
	ServerIP *int32 `json:"server_ip"`

	// IPv6 address of the Backend Server. Field introduced in 18.1.1.
	ServerIp6 *string `json:"server_ip6,omitempty"`

	// server_name of ConnectionLog.
	ServerName *string `json:"server_name,omitempty"`

	// Number of server_num_window_shrink.
	ServerNumWindowShrink *int32 `json:"server_num_window_shrink,omitempty"`

	// Number of server_out_of_orders.
	// Required: true
	ServerOutOfOrders *int32 `json:"server_out_of_orders"`

	// Number of server_retransmits.
	// Required: true
	ServerRetransmits *int32 `json:"server_retransmits"`

	//  Unit is MILLISECONDS.
	// Required: true
	ServerRtt *int32 `json:"server_rtt"`

	//  Unit is BYTES.
	// Required: true
	ServerRxBytes *int64 `json:"server_rx_bytes"`

	// Number of server_rx_pkts.
	// Required: true
	ServerRxPkts *int64 `json:"server_rx_pkts"`

	// Number of server_src_port.
	// Required: true
	ServerSrcPort *int32 `json:"server_src_port"`

	// Number of server_timeouts.
	// Required: true
	ServerTimeouts *int32 `json:"server_timeouts"`

	//  Unit is BYTES.
	// Required: true
	ServerTotalBytes *int64 `json:"server_total_bytes"`

	// Number of server_total_pkts.
	// Required: true
	ServerTotalPkts *int64 `json:"server_total_pkts"`

	//  Unit is BYTES.
	// Required: true
	ServerTxBytes *int64 `json:"server_tx_bytes"`

	// Number of server_tx_pkts.
	// Required: true
	ServerTxPkts *int64 `json:"server_tx_pkts"`

	// Number of server_zero_window_size_events.
	// Required: true
	ServerZeroWindowSizeEvents *int32 `json:"server_zero_window_size_events"`

	// service_engine of ConnectionLog.
	ServiceEngine *string `json:"service_engine,omitempty"`

	// significance of ConnectionLog.
	Significance *string `json:"significance,omitempty"`

	// Number of significant.
	// Required: true
	Significant *int64 `json:"significant"`

	// List of enums which indicate why a log is significant. Enum options - ADF_CLIENT_CONN_SETUP_REFUSED, ADF_SERVER_CONN_SETUP_REFUSED, ADF_CLIENT_CONN_SETUP_TIMEDOUT, ADF_SERVER_CONN_SETUP_TIMEDOUT, ADF_CLIENT_CONN_SETUP_FAILED_INTERNAL, ADF_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_CONN_SETUP_FAILED_BAD_PACKET, ADF_UDP_CONN_SETUP_FAILED_INTERNAL, ADF_UDP_SERVER_CONN_SETUP_FAILED_INTERNAL, ADF_CLIENT_SENT_RESET, ADF_SERVER_SENT_RESET, ADF_CLIENT_CONN_TIMEDOUT, ADF_SERVER_CONN_TIMEDOUT, ADF_USER_DELETE_OPERATION, ADF_CLIENT_REQUEST_TIMEOUT, ADF_CLIENT_CONN_ABORTED, ADF_CLIENT_SSL_HANDSHAKE_FAILURE, ADF_CLIENT_CONN_FAILED, ADF_SERVER_CERTIFICATE_VERIFICATION_FAILED, ADF_SERVER_SIDE_SSL_HANDSHAKE_FAILED...
	SignificantLog []string `json:"significant_log,omitempty"`

	// SIP related logging information. Field introduced in 17.2.12, 18.1.3, 18.2.1.
	SipLog *SipLog `json:"sip_log,omitempty"`

	//  Field introduced in 17.2.5.
	SniHostname *string `json:"sni_hostname,omitempty"`

	// ssl_cipher of ConnectionLog.
	SslCipher *string `json:"ssl_cipher,omitempty"`

	// ssl_session_id of ConnectionLog.
	SslSessionID *string `json:"ssl_session_id,omitempty"`

	// ssl_version of ConnectionLog.
	SslVersion *string `json:"ssl_version,omitempty"`

	// Number of start_timestamp.
	// Required: true
	StartTimestamp *int64 `json:"start_timestamp"`

	// Number of timeouts.
	// Required: true
	Timeouts *int32 `json:"timeouts"`

	//  Unit is BYTES.
	TotalBytes *int64 `json:"total_bytes,omitempty"`

	// Number of total_pkts.
	TotalPkts *int64 `json:"total_pkts,omitempty"`

	//  Unit is MILLISECONDS.
	TotalTime *int64 `json:"total_time,omitempty"`

	//  Unit is BYTES.
	// Required: true
	TxBytes *int64 `json:"tx_bytes"`

	// Number of tx_pkts.
	// Required: true
	TxPkts *int64 `json:"tx_pkts"`

	// Placeholder for description of property udf of obj type ConnectionLog field type str  type boolean
	// Required: true
	Udf *bool `json:"udf"`

	// Number of vcpu_id.
	// Required: true
	VcpuID *int32 `json:"vcpu_id"`

	// virtualservice of ConnectionLog.
	// Required: true
	Virtualservice *string `json:"virtualservice"`

	//  Field introduced in 17.1.1.
	VsIP *int32 `json:"vs_ip,omitempty"`

	// IPv6 address of the VIP of the VS. Field introduced in 18.1.1.
	VsIp6 *string `json:"vs_ip6,omitempty"`

	// Number of zero_window_size_events.
	// Required: true
	ZeroWindowSizeEvents *int32 `json:"zero_window_size_events"`
}

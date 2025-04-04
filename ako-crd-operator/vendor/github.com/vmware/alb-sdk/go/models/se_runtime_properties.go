// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeRuntimeProperties se runtime properties
// swagger:model SeRuntimeProperties
type SeRuntimeProperties struct {

	// Allow admin user ssh access to SE. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdminSSHEnabled *bool `json:"admin_ssh_enabled,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AppHeaders []*AppHdr `json:"app_headers,omitempty"`

	// Deprecated in 21.1.3. Use config in ServiceEngineGroup instead. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BaremetalDispatcherHandlesFlows *bool `json:"baremetal_dispatcher_handles_flows,omitempty"`

	// Rate limit on maximum adf lossy log to pushper second. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1000), Basic edition(Allowed values- 1000), Enterprise with Cloud Services edition.
	ConnectionsLossyLogRateLimiterThreshold *int32 `json:"connections_lossy_log_rate_limiter_threshold,omitempty"`

	// Rate limit on maximum adf udf or nf log to pushper second. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1000), Basic edition(Allowed values- 1000), Enterprise with Cloud Services edition.
	ConnectionsUdfnfLogRateLimiterThreshold *int32 `json:"connections_udfnf_log_rate_limiter_threshold,omitempty"`

	// Disable Flow Probes for Scaled out VS'es. (This field has been moved to se_group properties 20.1.3 onwards.). Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableFlowProbes *bool `json:"disable_flow_probes,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DosProfile *DosThresholdProfile `json:"dos_profile,omitempty"`

	// Timeout for downstream to become writable. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DownstreamSendTimeout *uint32 `json:"downstream_send_timeout,omitempty"`

	// Frequency of SE - SE HB messages when aggressive failure mode detection is enabled. (This field has been moved to se_group properties 20.1.3 onwards). Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 100), Basic edition(Allowed values- 100), Enterprise with Cloud Services edition.
	DpAggressiveHbFrequency *uint32 `json:"dp_aggressive_hb_frequency,omitempty"`

	// Consecutive HB failures after which failure is reported to controller,when aggressive failure mode detection is enabled. (This field has been moved to se_group properties 20.1.3 onwards). Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 10), Basic edition(Allowed values- 10), Enterprise with Cloud Services edition.
	DpAggressiveHbTimeoutCount *uint32 `json:"dp_aggressive_hb_timeout_count,omitempty"`

	// Frequency of SE - SE HB messages when aggressive failure mode detection is not enabled. (This field has been moved to se_group properties 20.1.3 onwards). Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DpHbFrequency *uint32 `json:"dp_hb_frequency,omitempty"`

	// Consecutive HB failures after which failure is reported to controller, when aggressive failure mode detection is not enabled. (This field has been moved to se_group properties 20.1.3 onwards). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DpHbTimeoutCount *uint32 `json:"dp_hb_timeout_count,omitempty"`

	// Frequency of ARP requests sent by SE for each VIP to detect duplicate IP when it loses conectivity to controller. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DupipFrequency *uint32 `json:"dupip_frequency,omitempty"`

	// Number of ARP responses received for the VIP after which SE decides that the VIP has been moved and disables the VIP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DupipTimeoutCount *uint32 `json:"dupip_timeout_count,omitempty"`

	// Enable HSM luna engine logs. Field introduced in 16.4.8, 17.1.11, 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableHsmLog *bool `json:"enable_hsm_log,omitempty"`

	// Enable proxy ARP from Host interface for Front End  proxies. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FeproxyVipsEnableProxyArp *bool `json:"feproxy_vips_enable_proxy_arp,omitempty"`

	// How often to push the flow table IPC messages in the main loop. The value is the number of times through the loop before pushing the batch. i.e, a value of 1 means every time through the loop. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowTableBatchPushFrequency *uint32 `json:"flow_table_batch_push_frequency,omitempty"`

	// Overrides the MTU value received via DHCP or some other means. Use this when the infrastructure advertises an MTU that is not supported by all devices in the network. For example, in AWS or when an overlay is used. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GlobalMtu *uint32 `json:"global_mtu,omitempty"`

	// Enable Javascript console logs on the client browser when collecting client insights. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	HTTPRumConsoleLog *bool `json:"http_rum_console_log,omitempty"`

	// Minimum response size content length to sample for client insights. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 64), Basic edition(Allowed values- 64), Enterprise with Cloud Services edition.
	HTTPRumMinContentLength *uint32 `json:"http_rum_min_content_length,omitempty"`

	// Number of requests to dispatch from the request queue at a regular interval. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LbactionNumRequestsToDispatch *uint32 `json:"lbaction_num_requests_to_dispatch,omitempty"`

	// Maximum retries per request in the request queue. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LbactionRqPerRequestMaxRetries *uint32 `json:"lbaction_rq_per_request_max_retries,omitempty"`

	// Deprecated in 21.1.1. Flag to indicate if log files are compressed upon full on the Service Engine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentCompressLogs *bool `json:"log_agent_compress_logs,omitempty"`

	// Deprecated in 21.1.1. Maximum application log file size before rollover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentFileSzAppl *uint32 `json:"log_agent_file_sz_appl,omitempty"`

	// Deprecated in 21.1.1. Maximum connection log file size before rollover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentFileSzConn *uint32 `json:"log_agent_file_sz_conn,omitempty"`

	// Deprecated in 21.1.1. Maximum debug log file size before rollover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentFileSzDebug *uint32 `json:"log_agent_file_sz_debug,omitempty"`

	// Deprecated in 21.1.1. Maximum event log file size before rollover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentFileSzEvent *uint32 `json:"log_agent_file_sz_event,omitempty"`

	//  Deprecated in 21.1.1. Minimum storage allocated for logs irrespective of memory and cores. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentLogStorageMinSz *uint32 `json:"log_agent_log_storage_min_sz,omitempty"`

	// Deprecated in 21.1.1. Maximum concurrent rsync requests initiated from log-agent to the Controller. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentMaxConcurrentRsync *uint32 `json:"log_agent_max_concurrent_rsync,omitempty"`

	// Deprecated in 21.1.1. Excess percentage threshold of disk size to trigger cleanup of logs on the Service Engine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentMaxStorageExcessPercent *uint32 `json:"log_agent_max_storage_excess_percent,omitempty"`

	// Deprecated in 21.1.1. Maximum storage on the disk not allocated for logs on the Service Engine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentMaxStorageIgnorePercent *float32 `json:"log_agent_max_storage_ignore_percent,omitempty"`

	// Deprecated in 21.1.1. Minimum storage allocated to any given VirtualService on the Service Engine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentMinStoragePerVs *uint32 `json:"log_agent_min_storage_per_vs,omitempty"`

	// Deprecated in 21.1.1. Internal timer to stall log-agent and prevent it from hogging CPU cycles on the Service Engine. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentSleepInterval *uint32 `json:"log_agent_sleep_interval,omitempty"`

	// Deprecated in 21.1.1. Timeout to purge unknown Virtual Service logs from the Service Engine. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentUnknownVsTimer *uint32 `json:"log_agent_unknown_vs_timer,omitempty"`

	// Deprecated in 21.1.1. Maximum number of file names in a log message. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogMessageMaxFileListSize *uint32 `json:"log_message_max_file_list_size,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NgxFreeConnectionStack *bool `json:"ngx_free_connection_stack,omitempty"`

	// Maximum memory in bytes allocated for persistence entries. Allowed values are 0-33554432. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PersistenceMemMax *uint32 `json:"persistence_mem_max,omitempty"`

	// Enable punting of UDP packets from primary to other Service Engines. This applies to Virtual Services with Per-Packet Loadbalancing enabled. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleoutUDPPerPkt *bool `json:"scaleout_udp_per_pkt,omitempty"`

	// LDAP basicauth default bind timeout enforced on connections to LDAP server. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeAuthLdapBindTimeout *uint32 `json:"se_auth_ldap_bind_timeout,omitempty"`

	// Size of LDAP basicauth credentials cache used on the dataplane. Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeAuthLdapCacheSize *uint32 `json:"se_auth_ldap_cache_size,omitempty"`

	// LDAP basicauth default connection timeout enforced on connections to LDAP server. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeAuthLdapConnectTimeout *uint32 `json:"se_auth_ldap_connect_timeout,omitempty"`

	// Number of concurrent connections to LDAP server by a single basic auth LDAP process. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeAuthLdapConnsPerServer *uint32 `json:"se_auth_ldap_conns_per_server,omitempty"`

	// LDAP basicauth default reconnect timeout enforced on connections to LDAP server. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeAuthLdapReconnectTimeout *uint32 `json:"se_auth_ldap_reconnect_timeout,omitempty"`

	// LDAP basicauth default login or group search request timeout enforced on connections to LDAP server. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeAuthLdapRequestTimeout *uint32 `json:"se_auth_ldap_request_timeout,omitempty"`

	// LDAP basicauth uses multiple ldap servers in the event of a failover only. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeAuthLdapServersFailoverOnly *bool `json:"se_auth_ldap_servers_failover_only,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDpCompression *SeRuntimeCompressionProperties `json:"se_dp_compression,omitempty"`

	// Deprecated - This field has been moved to se_group properties 20.1.3 onwards. Internal only. Used to simulate SE - SE HB failure. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpHmDrops *int32 `json:"se_dp_hm_drops,omitempty"`

	// Deprecated in 21.1.3. Use config in ServiceEngineGroup instead. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpIfStatePollInterval *uint32 `json:"se_dp_if_state_poll_interval,omitempty"`

	// Deprecated in 21.1.1. Internal buffer full indicator on the Service Engine beyond which the unfiltered logs are abandoned. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpLogNfEnqueuePercent *uint32 `json:"se_dp_log_nf_enqueue_percent,omitempty"`

	// Deprecated in 21.1.1. Internal buffer full indicator on the Service Engine beyond which the user filtered logs are abandoned. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpLogUdfEnqueuePercent *uint32 `json:"se_dp_log_udf_enqueue_percent,omitempty"`

	// Deprecated in 21.1.3. Use config in ServiceEngineGroup instead. Field introduced in 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDumpCoreOnAssert *bool `json:"se_dump_core_on_assert,omitempty"`

	// Accept/ignore interface routes (i.e, no next hop IP address). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHandleInterfaceRoutes *bool `json:"se_handle_interface_routes,omitempty"`

	// Internal use only. Allowed values are 0-7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHbPersistFudgeBits *uint32 `json:"se_hb_persist_fudge_bits,omitempty"`

	// Number of packets with wrong mac after which SE attempts to disable promiscious mode. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMacErrorThresholdToDisablePromiscious *uint32 `json:"se_mac_error_threshold_to_disable_promiscious,omitempty"`

	// Internal use only. Enables poisoning of freed memory blocks. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMemoryPoison *bool `json:"se_memory_poison,omitempty"`

	// Internal use only. Frequency (ms) of metrics updates from SE to controller. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 60000), Basic edition(Allowed values- 60000), Enterprise with Cloud Services edition.
	SeMetricsInterval *uint32 `json:"se_metrics_interval,omitempty"`

	// Internal use only. Enable or disable real time metrics irrespective of virtualservice or SE group configuration. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Essentials edition is false, Basic edition is false, Enterprise is True.
	SeMetricsRtEnabled *bool `json:"se_metrics_rt_enabled,omitempty"`

	// Internal use only. Frequency (ms) of realtime metrics updates from SE to controller. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMetricsRtInterval *uint32 `json:"se_metrics_rt_interval,omitempty"`

	// Deprecated in 21.1.3. Use config in ServiceEngineGroup instead. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePacketBufferMax *uint32 `json:"se_packet_buffer_max,omitempty"`

	// Internal use only. If enabled, randomly packets are dropped. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRandomTCPDrops *bool `json:"se_random_tcp_drops,omitempty"`

	// SE rate limiters. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeRateLimiters *SeRateLimiters `json:"se_rate_limiters,omitempty"`

	// IP ranges on which there may be virtual services (for configuring iptables/routes). Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceIPSubnets []*IPAddrPrefix `json:"service_ip_subnets,omitempty"`

	// Port ranges on which there may be virtual services (for configuring iptables). Applicable in container ecosystems like Mesos. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServicePortRanges []*PortRange `json:"service_port_ranges,omitempty"`

	// Make service ports accessible on all Host interfaces in addition to East-West VIP and/or bridge IP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServicesAccessibleAllInterfaces *bool `json:"services_accessible_all_interfaces,omitempty"`

	// Default value for max number of retransmissions for a SYN packet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPSyncacheMaxRetransmitDefault *uint32 `json:"tcp_syncache_max_retransmit_default,omitempty"`

	// Timeout for backend connection. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamConnectTimeout *uint32 `json:"upstream_connect_timeout,omitempty"`

	// L7 Upstream Connection pool cache threshold in percentage. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamConnpoolCacheThresh *int32 `json:"upstream_connpool_cache_thresh,omitempty"`

	// Idle timeout value for a connection in the upstream connection pool, when the current cache size is above the threshold. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamConnpoolConnIDLEThreshTmo *int32 `json:"upstream_connpool_conn_idle_thresh_tmo,omitempty"`

	// L7 Upstream Connection pool max cache size per core. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamConnpoolCoreMaxCache *int32 `json:"upstream_connpool_core_max_cache,omitempty"`

	// Enable upstream connection pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamConnpoolEnable *bool `json:"upstream_connpool_enable,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamKeepalive *bool `json:"upstream_keepalive,omitempty"`

	// Timeout for data to be received from backend. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpstreamReadTimeout *uint32 `json:"upstream_read_timeout,omitempty"`

	// Timeout for upstream to become writable. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 3600000), Basic edition(Allowed values- 3600000), Enterprise with Cloud Services edition.
	UpstreamSendTimeout *uint32 `json:"upstream_send_timeout,omitempty"`

	// Defines in seconds how long before an unused user-defined-metric is garbage collected. Field introduced in 17.1.5. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserDefinedMetricAge *uint32 `json:"user_defined_metric_age,omitempty"`
}

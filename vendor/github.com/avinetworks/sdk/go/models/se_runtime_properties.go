package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeRuntimeProperties se runtime properties
// swagger:model SeRuntimeProperties
type SeRuntimeProperties struct {

	// Allow admin user ssh access to SE. Field introduced in 18.2.5.
	AdminSSHEnabled *bool `json:"admin_ssh_enabled,omitempty"`

	// Placeholder for description of property app_headers of obj type SeRuntimeProperties field type str  type object
	AppHeaders []*AppHdr `json:"app_headers,omitempty"`

	// Control if dispatcher core also handles TCP flows in baremetal SE.
	BaremetalDispatcherHandlesFlows *bool `json:"baremetal_dispatcher_handles_flows,omitempty"`

	// Rate limit on maximum adf lossy log to pushper second.
	ConnectionsLossyLogRateLimiterThreshold *int32 `json:"connections_lossy_log_rate_limiter_threshold,omitempty"`

	// Rate limit on maximum adf udf or nf log to pushper second.
	ConnectionsUdfnfLogRateLimiterThreshold *int32 `json:"connections_udfnf_log_rate_limiter_threshold,omitempty"`

	// Disable Flow Probes for Scaled out VS'es. Field introduced in 17.1.1.
	DisableFlowProbes *bool `json:"disable_flow_probes,omitempty"`

	// Deprecated. Field deprecated in 17.2.5. Field introduced in 17.2.1.
	DisableGro *bool `json:"disable_gro,omitempty"`

	// Deprecated. Field deprecated in 17.2.5. Field introduced in 17.2.4.
	DisableTso *bool `json:"disable_tso,omitempty"`

	// Placeholder for description of property dos_profile of obj type SeRuntimeProperties field type str  type object
	DosProfile *DosThresholdProfile `json:"dos_profile,omitempty"`

	// Timeout for downstream to become writable.
	DownstreamSendTimeout *int32 `json:"downstream_send_timeout,omitempty"`

	// Frequency of SE - SE HB messages when aggressive failure mode detection is enabled.
	DpAggressiveHbFrequency *int32 `json:"dp_aggressive_hb_frequency,omitempty"`

	// Consecutive HB failures after which failure is reported to controller,when aggressive failure mode detection is enabled.
	DpAggressiveHbTimeoutCount *int32 `json:"dp_aggressive_hb_timeout_count,omitempty"`

	// Frequency of SE - SE HB messages when aggressive failure mode detection is not enabled.
	DpHbFrequency *int32 `json:"dp_hb_frequency,omitempty"`

	// Consecutive HB failures after which failure is reported to controller, when aggressive failure mode detection is not enabled.
	DpHbTimeoutCount *int32 `json:"dp_hb_timeout_count,omitempty"`

	// Frequency of ARP requests sent by SE for each VIP to detect duplicate IP when it loses conectivity to controller.
	DupipFrequency *int32 `json:"dupip_frequency,omitempty"`

	// Number of ARP responses received for the VIP after which SE decides that the VIP has been moved and disables the VIP.
	DupipTimeoutCount *int32 `json:"dupip_timeout_count,omitempty"`

	// Enable HSM luna engine logs. Field introduced in 16.4.8, 17.1.11, 17.2.3.
	EnableHsmLog *bool `json:"enable_hsm_log,omitempty"`

	// Enable proxy ARP from Host interface for Front End  proxies.
	FeproxyVipsEnableProxyArp *bool `json:"feproxy_vips_enable_proxy_arp,omitempty"`

	// How often to push the flow table IPC messages in the main loop. The value is the number of times through the loop before pushing the batch. i.e, a value of 1 means every time through the loop.
	FlowTableBatchPushFrequency *int32 `json:"flow_table_batch_push_frequency,omitempty"`

	// Deprecated. Field deprecated in 17.1.1.
	FlowTableMaxEntriesDeprecated *int32 `json:"flow_table_max_entries_deprecated,omitempty"`

	// Deprecated. Field deprecated in 17.2.5.
	FlowTableNewSynMaxEntries *int32 `json:"flow_table_new_syn_max_entries,omitempty"`

	// Overrides the MTU value received via DHCP or some other means. Use this when the infrastructure advertises an MTU that is not supported by all devices in the network. For example, in AWS or when an overlay is used.
	GlobalMtu *int32 `json:"global_mtu,omitempty"`

	// Enable Javascript console logs on the client browser when collecting client insights.
	HTTPRumConsoleLog *bool `json:"http_rum_console_log,omitempty"`

	// Minimum response size content length to sample for client insights.
	HTTPRumMinContentLength *int32 `json:"http_rum_min_content_length,omitempty"`

	// How often to push the LB IPC messages in the main loop. The value is the number of times the loop has to run before pushing the batch. i.e, a value of 1 means the batch is pushed every time the loop runs. Field deprecated in 18.1.3. Field introduced in 17.2.8.
	LbBatchPushFrequency *int32 `json:"lb_batch_push_frequency,omitempty"`

	// Deprecated. Field deprecated in 17.1.1.
	LbFailMaxTime *int32 `json:"lb_fail_max_time,omitempty"`

	// Number of requests to dispatch from the request queue at a regular interval.
	LbactionNumRequestsToDispatch *int32 `json:"lbaction_num_requests_to_dispatch,omitempty"`

	// Maximum retries per request in the request queue.
	LbactionRqPerRequestMaxRetries *int32 `json:"lbaction_rq_per_request_max_retries,omitempty"`

	// Flag to indicate if log files are compressed upon full on the Service Engine.
	LogAgentCompressLogs *bool `json:"log_agent_compress_logs,omitempty"`

	// Log-agent test property used to simulate slow TCP connections.
	LogAgentConnSendBufferSize *int32 `json:"log_agent_conn_send_buffer_size,omitempty"`

	// Maximum size of data sent by log-agent to Controller over the TCP connection.
	LogAgentExportMsgBufferSize *int32 `json:"log_agent_export_msg_buffer_size,omitempty"`

	// Time log-agent waits before sending data to the Controller.
	LogAgentExportWaitTime *int32 `json:"log_agent_export_wait_time,omitempty"`

	// Maximum application log file size before rollover.
	LogAgentFileSzAppl *int32 `json:"log_agent_file_sz_appl,omitempty"`

	// Maximum connection log file size before rollover.
	LogAgentFileSzConn *int32 `json:"log_agent_file_sz_conn,omitempty"`

	// Maximum debug log file size before rollover.
	LogAgentFileSzDebug *int32 `json:"log_agent_file_sz_debug,omitempty"`

	// Maximum event log file size before rollover.
	LogAgentFileSzEvent *int32 `json:"log_agent_file_sz_event,omitempty"`

	// Minimum storage allocated for logs irrespective of memory and cores.
	LogAgentLogStorageMinSz *int32 `json:"log_agent_log_storage_min_sz,omitempty"`

	// Maximum number of Virtual Service log files maintained for significant logs on the Service Engine.
	LogAgentMaxActiveAdfFilesPerVs *int32 `json:"log_agent_max_active_adf_files_per_vs,omitempty"`

	// Maximum concurrent rsync requests initiated from log-agent to the Controller.
	LogAgentMaxConcurrentRsync *int32 `json:"log_agent_max_concurrent_rsync,omitempty"`

	// Maximum size of serialized log message on the Service Engine.
	LogAgentMaxLogmessageProtoSz *int32 `json:"log_agent_max_logmessage_proto_sz,omitempty"`

	// Excess percentage threshold of disk size to trigger cleanup of logs on the Service Engine.
	LogAgentMaxStorageExcessPercent *int32 `json:"log_agent_max_storage_excess_percent,omitempty"`

	// Maximum storage on the disk not allocated for logs on the Service Engine.
	LogAgentMaxStorageIgnorePercent *float32 `json:"log_agent_max_storage_ignore_percent,omitempty"`

	// Minimum storage allocated to any given VirtualService on the Service Engine.
	LogAgentMinStoragePerVs *int32 `json:"log_agent_min_storage_per_vs,omitempty"`

	// Time interval log-agent pauses between logs obtained from the dataplane.
	LogAgentPauseInterval *int32 `json:"log_agent_pause_interval,omitempty"`

	// Internal timer to stall log-agent and prevent it from hogging CPU cycles on the Service Engine.
	LogAgentSleepInterval *int32 `json:"log_agent_sleep_interval,omitempty"`

	// Timeout to purge unknown Virtual Service logs from the Service Engine.
	LogAgentUnknownVsTimer *int32 `json:"log_agent_unknown_vs_timer,omitempty"`

	// Maximum number of file names in a log message.
	LogMessageMaxFileListSize *int32 `json:"log_message_max_file_list_size,omitempty"`

	// Deprecated. Field deprecated in 17.1.1.
	MaxThroughput *int32 `json:"max_throughput,omitempty"`

	// enables mcache - controls fetch/store/store_out.
	McacheEnabled *bool `json:"mcache_enabled,omitempty"`

	// enables mcache_fetch.
	McacheFetchEnabled *bool `json:"mcache_fetch_enabled,omitempty"`

	// Use SE Group's app_cache_percent to set cache memory usage limit on SE. Field deprecated in 18.2.3.
	McacheMaxCacheSize *int64 `json:"mcache_max_cache_size,omitempty"`

	// enables mcache_store.
	McacheStoreInEnabled *bool `json:"mcache_store_in_enabled,omitempty"`

	// max object size.
	McacheStoreInMaxSize *int32 `json:"mcache_store_in_max_size,omitempty"`

	// min object size.
	McacheStoreInMinSize *int32 `json:"mcache_store_in_min_size,omitempty"`

	// enables mcache_store_out.
	McacheStoreOutEnabled *bool `json:"mcache_store_out_enabled,omitempty"`

	// Use SE Group's app_cache_percent to set cache memory usage limit on SE. Field deprecated in 18.2.3.
	McacheStoreSeMaxSize *int64 `json:"mcache_store_se_max_size,omitempty"`

	// Placeholder for description of property ngx_free_connection_stack of obj type SeRuntimeProperties field type str  type boolean
	NgxFreeConnectionStack *bool `json:"ngx_free_connection_stack,omitempty"`

	// Deprecated. Field deprecated in 17.1.1.
	PersistenceEntriesLowWatermark *int32 `json:"persistence_entries_low_watermark,omitempty"`

	// Maximum memory in bytes allocated for persistence entries. Allowed values are 0-33554432.
	PersistenceMemMax *int32 `json:"persistence_mem_max,omitempty"`

	// Enable punting of UDP packets from primary to other Service Engines. This applies to Virtual Services with Per-Packet Loadbalancing enabled.
	ScaleoutUDPPerPkt *bool `json:"scaleout_udp_per_pkt,omitempty"`

	// LDAP basicauth default bind timeout enforced on connections to LDAP server.
	SeAuthLdapBindTimeout *int32 `json:"se_auth_ldap_bind_timeout,omitempty"`

	// Size of LDAP basicauth credentials cache used on the dataplane.
	SeAuthLdapCacheSize *int32 `json:"se_auth_ldap_cache_size,omitempty"`

	// LDAP basicauth default connection timeout enforced on connections to LDAP server.
	SeAuthLdapConnectTimeout *int32 `json:"se_auth_ldap_connect_timeout,omitempty"`

	// Number of concurrent connections to LDAP server by a single basic auth LDAP process.
	SeAuthLdapConnsPerServer *int32 `json:"se_auth_ldap_conns_per_server,omitempty"`

	// LDAP basicauth default reconnect timeout enforced on connections to LDAP server.
	SeAuthLdapReconnectTimeout *int32 `json:"se_auth_ldap_reconnect_timeout,omitempty"`

	// LDAP basicauth default login or group search request timeout enforced on connections to LDAP server.
	SeAuthLdapRequestTimeout *int32 `json:"se_auth_ldap_request_timeout,omitempty"`

	// LDAP basicauth uses multiple ldap servers in the event of a failover only.
	SeAuthLdapServersFailoverOnly *bool `json:"se_auth_ldap_servers_failover_only,omitempty"`

	// Placeholder for description of property se_dp_compression of obj type SeRuntimeProperties field type str  type object
	SeDpCompression *SeRuntimeCompressionProperties `json:"se_dp_compression,omitempty"`

	// Internal only. Used to simulate SE - SE HB failure.
	SeDpHmDrops *int32 `json:"se_dp_hm_drops,omitempty"`

	// Number of jiffies between polling interface state.
	SeDpIfStatePollInterval *int32 `json:"se_dp_if_state_poll_interval,omitempty"`

	// Internal buffer full indicator on the Service Engine beyond which the unfiltered logs are abandoned.
	SeDpLogNfEnqueuePercent *int32 `json:"se_dp_log_nf_enqueue_percent,omitempty"`

	// Internal buffer full indicator on the Service Engine beyond which the user filtered logs are abandoned.
	SeDpLogUdfEnqueuePercent *int32 `json:"se_dp_log_udf_enqueue_percent,omitempty"`

	// Deprecated. Field deprecated in 18.2.5. Field introduced in 17.1.1.
	SeDpVnicQueueStallEventSleep *int32 `json:"se_dp_vnic_queue_stall_event_sleep,omitempty"`

	// Deprecated. Field deprecated in 18.2.5. Field introduced in 17.1.1.
	SeDpVnicQueueStallThreshold *int32 `json:"se_dp_vnic_queue_stall_threshold,omitempty"`

	// Deprecated. Field deprecated in 18.2.5. Field introduced in 17.1.1.
	SeDpVnicQueueStallTimeout *int32 `json:"se_dp_vnic_queue_stall_timeout,omitempty"`

	// Deprecated. Field deprecated in 18.2.5. Field introduced in 17.1.14, 17.2.5, 18.1.1.
	SeDpVnicRestartOnQueueStallCount *int32 `json:"se_dp_vnic_restart_on_queue_stall_count,omitempty"`

	// Deprecated. Field deprecated in 18.2.5. Field introduced in 17.1.14, 17.2.5, 18.1.1.
	SeDpVnicStallSeRestartWindow *int32 `json:"se_dp_vnic_stall_se_restart_window,omitempty"`

	// Enable core dump on assert. Field introduced in 18.1.3, 18.2.1.
	SeDumpCoreOnAssert *bool `json:"se_dump_core_on_assert,omitempty"`

	// Accept/ignore interface routes (i.e, no next hop IP address).
	SeHandleInterfaceRoutes *bool `json:"se_handle_interface_routes,omitempty"`

	// Internal use only. Allowed values are 0-7.
	SeHbPersistFudgeBits *int32 `json:"se_hb_persist_fudge_bits,omitempty"`

	// Number of packets with wrong mac after which SE attempts to disable promiscious mode.
	SeMacErrorThresholdToDisablePromiscious *int32 `json:"se_mac_error_threshold_to_disable_promiscious,omitempty"`

	// Deprecated. Field deprecated in 17.1.1.
	SeMallocThresh *int32 `json:"se_malloc_thresh,omitempty"`

	// Internal use only. Enables poisoning of freed memory blocks.
	SeMemoryPoison *bool `json:"se_memory_poison,omitempty"`

	// Internal use only. Frequency (ms) of metrics updates from SE to controller.
	SeMetricsInterval *int32 `json:"se_metrics_interval,omitempty"`

	// Internal use only. Enable or disable real time metrics irrespective of virtualservice or SE group configuration.
	SeMetricsRtEnabled *bool `json:"se_metrics_rt_enabled,omitempty"`

	// Internal use only. Frequency (ms) of realtime metrics updates from SE to controller.
	SeMetricsRtInterval *int32 `json:"se_metrics_rt_interval,omitempty"`

	// Internal use only. Used to artificially reduce the available number of packet buffers.
	SePacketBufferMax *int32 `json:"se_packet_buffer_max,omitempty"`

	// Internal use only. If enabled, randomly packets are dropped.
	SeRandomTCPDrops *bool `json:"se_random_tcp_drops,omitempty"`

	// SE rate limiters.
	SeRateLimiters *SeRateLimiters `json:"se_rate_limiters,omitempty"`

	// IP ranges on which there may be virtual services (for configuring iptables/routes).
	ServiceIPSubnets []*IPAddrPrefix `json:"service_ip_subnets,omitempty"`

	// Port ranges on which there may be virtual services (for configuring iptables). Applicable in container ecosystems like Mesos.
	ServicePortRanges []*PortRange `json:"service_port_ranges,omitempty"`

	// Make service ports accessible on all Host interfaces in addition to East-West VIP and/or bridge IP.
	ServicesAccessibleAllInterfaces *bool `json:"services_accessible_all_interfaces,omitempty"`

	// Placeholder for description of property spdy_fwd_proxy_parse_enable of obj type SeRuntimeProperties field type str  type boolean
	SpdyFwdProxyParseEnable *bool `json:"spdy_fwd_proxy_parse_enable,omitempty"`

	// Maximum size of the SYN cache table. After this limit is reached, SYN cookies are used. This is per core of the serviceengine. Field deprecated in 17.2.5.
	TCPSynCacheMax *int32 `json:"tcp_syn_cache_max,omitempty"`

	// Default value for max number of retransmissions for a SYN packet.
	TCPSyncacheMaxRetransmitDefault *int32 `json:"tcp_syncache_max_retransmit_default,omitempty"`

	// Timeout for backend connection.
	UpstreamConnectTimeout *int32 `json:"upstream_connect_timeout,omitempty"`

	// L7 Upstream Connection pool cache threshold in percentage.
	UpstreamConnpoolCacheThresh *int32 `json:"upstream_connpool_cache_thresh,omitempty"`

	// Idle timeout value for a connection in the upstream connection pool, when the current cache size is above the threshold.
	UpstreamConnpoolConnIDLEThreshTmo *int32 `json:"upstream_connpool_conn_idle_thresh_tmo,omitempty"`

	// Deprecated. Field deprecated in 18.2.1.
	UpstreamConnpoolConnIDLETmo *int32 `json:"upstream_connpool_conn_idle_tmo,omitempty"`

	// Deprecated. Field deprecated in 18.2.1.
	UpstreamConnpoolConnLifeTmo *int32 `json:"upstream_connpool_conn_life_tmo,omitempty"`

	// Deprecated. Field deprecated in 18.2.1.
	UpstreamConnpoolConnMaxReuse *int32 `json:"upstream_connpool_conn_max_reuse,omitempty"`

	// L7 Upstream Connection pool max cache size per core.
	UpstreamConnpoolCoreMaxCache *int32 `json:"upstream_connpool_core_max_cache,omitempty"`

	// Enable upstream connection pool.
	UpstreamConnpoolEnable *bool `json:"upstream_connpool_enable,omitempty"`

	// Deprecated. Field deprecated in 18.2.1.
	UpstreamConnpoolServerMaxCache *int32 `json:"upstream_connpool_server_max_cache,omitempty"`

	// Number of upstream_connpool_strategy.
	UpstreamConnpoolStrategy *int32 `json:"upstream_connpool_strategy,omitempty"`

	// Placeholder for description of property upstream_keepalive of obj type SeRuntimeProperties field type str  type boolean
	UpstreamKeepalive *bool `json:"upstream_keepalive,omitempty"`

	// Timeout for data to be received from backend.
	UpstreamReadTimeout *int32 `json:"upstream_read_timeout,omitempty"`

	// Timeout for upstream to become writable.
	UpstreamSendTimeout *int32 `json:"upstream_send_timeout,omitempty"`

	// Defines in seconds how long before an unused user-defined-metric is garbage collected. Field introduced in 17.1.5.
	UserDefinedMetricAge *int32 `json:"user_defined_metric_age,omitempty"`
}

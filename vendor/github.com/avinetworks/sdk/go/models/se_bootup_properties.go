package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeBootupProperties se bootup properties
// swagger:model SeBootupProperties
type SeBootupProperties struct {

	// Deprecated. Field deprecated in 17.2.8. Field introduced in 17.1.1.
	DistributeQueues *bool `json:"distribute_queues,omitempty"`

	// Deprecated. Field deprecated in 18.2.5.
	DistributeVnics *bool `json:"distribute_vnics,omitempty"`

	// End of the Local TCP port range used by SE for backend connections in docker environment.
	DockerBackendPortend *int32 `json:"docker_backend_portend,omitempty"`

	// Start of the Local TCP port range used by SE for backend connections in docker environment.
	DockerBackendPortstart *int32 `json:"docker_backend_portstart,omitempty"`

	// Enable or disable fair queueing for packet transmission among virtualservices on an SE.
	FairQueueingEnabled *bool `json:"fair_queueing_enabled,omitempty"`

	// Granularity or Resolution of co-ordinates used. When the value is 1 the co-ordinates provided in the geo-db are used as is (highest resolution.This value provides a 'zoom-out' value so that coarser co-ordinates are used. With higher resolution, logs can contain finer location information. But, lower resolution provides significant memory and cpu benefits on the service engine. Besides, given a smaller number of members that are separated geographically, a lower resolution is sufficient for correct load-balancing. Allowed values are 1-20. Field introduced in 17.1.1.
	GeoDbGranularity *int32 `json:"geo_db_granularity,omitempty"`

	// Number of L7 connections that can be cached per core.
	L7ConnsPerCore *int32 `json:"l7_conns_per_core,omitempty"`

	// Number of reserved L7 listener connections per core.
	L7ResvdListenConnsPerCore *int32 `json:"l7_resvd_listen_conns_per_core,omitempty"`

	// Enable debug logs by default on Service Engine. This includes all other debugging logs. Debug logs can also be explcitly enabled from the CLI shell.
	LogAgentDebugEnabled *bool `json:"log_agent_debug_enabled,omitempty"`

	// Enable trace logs by default on Service Engine. Configuration operations are logged along with other important logs by Service Engine.
	LogAgentTraceEnabled *bool `json:"log_agent_trace_enabled,omitempty"`

	// Placeholder for description of property se_dp_compression of obj type SeBootupProperties field type str  type object
	SeDpCompression *SeBootupCompressionProperties `json:"se_dp_compression,omitempty"`

	// This field has been moved to se_group properties 18.1.2 onwards. Field deprecated in 18.1.3.
	SeDpdkPmd *int32 `json:"se_dpdk_pmd,omitempty"`

	// Use this to emulate more/less cpus than is actually available. One datapath process is started for each core.
	SeEmulatedCores *int32 `json:"se_emulated_cores,omitempty"`

	// Determines if SE-SE IPC messages are encapsulated in an IP header       0        Automatically determine based on hypervisor type    1        Use IP encap unconditionally    ~[0,1]   Don't use IP encap.
	SeIPEncapIpc *int32 `json:"se_ip_encap_ipc,omitempty"`

	// Determines if SE-SE IPC messages use SE interface IP instead of VIP        0        Automatically determine based on hypervisor type    1        Use SE interface IP unconditionally    ~[0,1]   Don't use SE interface IP.
	SeL3EncapIpc *int32 `json:"se_l3_encap_ipc,omitempty"`

	// Internal flag that blocks dataplane until all application logs are flushed to log-agent process.
	SeLogBufferAppBlockingDequeue *bool `json:"se_log_buffer_app_blocking_dequeue,omitempty"`

	// Internal application log buffer size to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments.
	SeLogBufferApplogSize *int32 `json:"se_log_buffer_applog_size,omitempty"`

	// Number of internal buffer chunks to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments.
	SeLogBufferChunkCount *int32 `json:"se_log_buffer_chunk_count,omitempty"`

	// Internal flag that blocks dataplane until all connection logs are flushed to log-agent process.
	SeLogBufferConnBlockingDequeue *bool `json:"se_log_buffer_conn_blocking_dequeue,omitempty"`

	// Internal connection log buffer size to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments.
	SeLogBufferConnlogSize *int32 `json:"se_log_buffer_connlog_size,omitempty"`

	// Internal flag that blocks dataplane until all outstanding events are flushed to log-agent process.
	SeLogBufferEventsBlockingDequeue *bool `json:"se_log_buffer_events_blocking_dequeue,omitempty"`

	// Internal events buffer size to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments.
	SeLogBufferEventsSize *int32 `json:"se_log_buffer_events_size,omitempty"`

	// Deprecated. Field deprecated in 18.2.5.
	SeLro *int32 `json:"se_lro,omitempty"`

	// Deprecated. Field deprecated in 18.2.5.
	SePcapPktCount *int32 `json:"se_pcap_pkt_count,omitempty"`

	// Deprecated. Field deprecated in 18.2.5.
	SePcapPktSz *int32 `json:"se_pcap_pkt_sz,omitempty"`

	// Deprecated. Field deprecated in 18.2.6.
	SeRumSamplingNavInterval *int32 `json:"se_rum_sampling_nav_interval,omitempty"`

	// Deprecated. Field deprecated in 18.2.6.
	SeRumSamplingNavPercent *int32 `json:"se_rum_sampling_nav_percent,omitempty"`

	// Deprecated. Field deprecated in 18.2.6.
	SeRumSamplingResInterval *int32 `json:"se_rum_sampling_res_interval,omitempty"`

	// Deprecated. Field deprecated in 18.2.6.
	SeRumSamplingResPercent *int32 `json:"se_rum_sampling_res_percent,omitempty"`

	// Determines if DSR from secondary SE is active or not      0        Automatically determine based on hypervisor type    1        Disable DSR unconditionally    ~[0,1]   Enable DSR unconditionally. Field deprecated in 17.1.1.
	SeTunnelMode *int32 `json:"se_tunnel_mode,omitempty"`

	// Deprecated. Field deprecated in 18.2.5.
	SeTxBatchSize *int32 `json:"se_tx_batch_size,omitempty"`

	// This field has been moved to se_group properties 18.1.2 onwards. Field deprecated in 18.1.3.
	SeUseDpdk *int32 `json:"se_use_dpdk,omitempty"`

	// Number of ssl_sess_cache_per_vs.
	SslSessCachePerVs *int32 `json:"ssl_sess_cache_per_vs,omitempty"`

	// Number of ssl_sess_cache_timeout.
	SslSessCacheTimeout *int32 `json:"ssl_sess_cache_timeout,omitempty"`

	// Size of the TCP SYN cache hash table.
	TCPSyncacheHashsize *int32 `json:"tcp_syncache_hashsize,omitempty"`
}

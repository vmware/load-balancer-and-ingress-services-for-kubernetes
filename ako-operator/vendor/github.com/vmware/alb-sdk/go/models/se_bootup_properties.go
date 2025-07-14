// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeBootupProperties se bootup properties
// swagger:model SeBootupProperties
type SeBootupProperties struct {

	// End of the Local TCP port range used by SE for backend connections in docker environment. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerBackendPortend *uint32 `json:"docker_backend_portend,omitempty"`

	// Start of the Local TCP port range used by SE for backend connections in docker environment. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerBackendPortstart *uint32 `json:"docker_backend_portstart,omitempty"`

	// Enable or disable fair queueing for packet transmission among virtualservices on an SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FairQueueingEnabled *bool `json:"fair_queueing_enabled,omitempty"`

	// Granularity or Resolution of co-ordinates used. When the value is 1 the co-ordinates provided in the geo-db are used as is (highest resolution.This value provides a 'zoom-out' value so that coarser co-ordinates are used. With higher resolution, logs can contain finer location information. But, lower resolution provides significant memory and cpu benefits on the service engine. Besides, given a smaller number of members that are separated geographically, a lower resolution is sufficient for correct load-balancing. Allowed values are 1-20. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1), Basic edition(Allowed values- 1), Enterprise with Cloud Services edition.
	GeoDbGranularity *uint32 `json:"geo_db_granularity,omitempty"`

	// Number of L7 connections that can be cached per core. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	L7ConnsPerCore *uint32 `json:"l7_conns_per_core,omitempty"`

	// Number of reserved L7 listener connections per core. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	L7ResvdListenConnsPerCore *uint32 `json:"l7_resvd_listen_conns_per_core,omitempty"`

	// Deprecated in 21.1.1. Enable debug logs by default on Service Engine. This includes all other debugging logs. Debug logs can also be explcitly enabled from the CLI shell. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentDebugEnabled *bool `json:"log_agent_debug_enabled,omitempty"`

	// Deprecated in 21.1.1. Enable trace logs by default on Service Engine. Configuration operations are logged along with other important logs by Service Engine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentTraceEnabled *bool `json:"log_agent_trace_enabled,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDpCompression *SeBootupCompressionProperties `json:"se_dp_compression,omitempty"`

	// Deprecated in 21.1.3. Use config in ServiceEngineGroup instead. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	SeEmulatedCores *uint32 `json:"se_emulated_cores,omitempty"`

	// Determines if SE-SE IPC messages are encapsulated in an IP header   Note  This field has been moved to se_group properties 20.1.3 onwards.    0        Automatically determine based on hypervisor type    1        Use IP encap unconditionally    ~[0,1]   Don't use IP encap. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIPEncapIpc *uint32 `json:"se_ip_encap_ipc,omitempty"`

	// Determines if SE-SE IPC messages use SE interface IP instead of VIP    Note  This field has been moved to se_group properties 20.1.3 onwards.    0        Automatically determine based on hypervisor type    1        Use SE interface IP unconditionally    ~[0,1]   Don't use SE interface IP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeL3EncapIpc *uint32 `json:"se_l3_encap_ipc,omitempty"`

	// Deprecated in 21.1.1. Internal flag that blocks dataplane until all application logs are flushed to log-agent process. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLogBufferAppBlockingDequeue *bool `json:"se_log_buffer_app_blocking_dequeue,omitempty"`

	// Internal application log buffer size to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments. Unit is WORD. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLogBufferApplogSize *uint32 `json:"se_log_buffer_applog_size,omitempty"`

	// Number of internal buffer chunks to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments. Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLogBufferChunkCount *uint32 `json:"se_log_buffer_chunk_count,omitempty"`

	// Deprecated in 21.1.1. Internal flag that blocks dataplane until all connection logs are flushed to log-agent process. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLogBufferConnBlockingDequeue *bool `json:"se_log_buffer_conn_blocking_dequeue,omitempty"`

	// Internal connection log buffer size to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments. Unit is WORD. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLogBufferConnlogSize *uint32 `json:"se_log_buffer_connlog_size,omitempty"`

	// Deprecated in 21.1.1. Internal flag that blocks dataplane until all outstanding events are flushed to log-agent process. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLogBufferEventsBlockingDequeue *bool `json:"se_log_buffer_events_blocking_dequeue,omitempty"`

	// Internal events buffer size to use on Service Engine. Can be fine tuned for better performance of data plane in specific environments. Unit is WORD. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLogBufferEventsSize *uint32 `json:"se_log_buffer_events_size,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslSessCachePerVs *uint32 `json:"ssl_sess_cache_per_vs,omitempty"`

	// Size of the TCP SYN cache hash table. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPSyncacheHashsize *uint32 `json:"tcp_syncache_hashsize,omitempty"`
}

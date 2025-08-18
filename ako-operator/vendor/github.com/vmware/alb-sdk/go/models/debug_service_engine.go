// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugServiceEngine debug service engine
// swagger:model DebugServiceEngine
type DebugServiceEngine struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Action to be invoked at configured layer. Enum options - SE_BENCHMARK_MODE_DROP, SE_BENCHMARK_MODE_REFLECT. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BenchmarkAction *string `json:"benchmark_action,omitempty"`

	// Toggle and configure the layer to benchmark performance. This can be done at a specific point in the SE packet processing pipeline. Enum options - SE_BENCHMARK_LAYER_NONE, SE_BENCHMARK_LAYER_POST_VNIC_RX, SE_BENCHMARK_LAYER_POST_FT_LOOKUP, SE_BENCHMARK_LAYER_NSP_LOOKUP, SE_BENCHMARK_LAYER_PRE_PROXY_PUNT, SE_BENCHMARK_LAYER_POST_PROXY_PUNT, SE_BENCHMARK_LAYER_ETHER_INPUT, SE_BENCHMARK_LAYER_IP_INPUT, SE_BENCHMARK_LAYER_UDP_INPUT, SE_BENCHMARK_LAYER_POST_L2_PROCESSING, SE_BENCHMARK_LAYER_POST_BUILD_KEY_LITE. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BenchmarkLayer *string `json:"benchmark_layer,omitempty"`

	// Configure different reflect modes. Enum options - SE_BENCHMARK_REFLECT_SWAP_L4, SE_BENCHMARK_REFLECT_SWAP_L2, SE_BENCHMARK_REFLECT_SWAP_L3. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BenchmarkOption *string `json:"benchmark_option,omitempty"`

	// RSS Hash function to be used for packet reflect in TX path. Enum options - SE_BENCHMARK_DISABLE_HASH, SE_BENCHMARK_RTE_SOFT_HASH. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BenchmarkRssHash *string `json:"benchmark_rss_hash,omitempty"`

	// Enable/disable packet capture. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Capture *bool `json:"capture,omitempty"`

	// Per packet capture filters for Debug Service Engine. Not applicable for DOS pcap capture. . Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaptureFilters *CaptureFilters `json:"capture_filters,omitempty"`

	// Params for SE pcap. Field introduced in 17.2.14,18.1.5,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaptureParams *DebugVirtualServiceCapture `json:"capture_params,omitempty"`

	// Per packet capture filters for Debug Service Engine. Not applicable for DOS pcap capture. . Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CapturePktFilter *CapturePacketFilter `json:"capture_pkt_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CPUShares []*DebugSeCPUShares `json:"cpu_shares,omitempty"`

	// Per packet IP filter for Service Engine PCAP. Matches with source and destination address. Field introduced in 17.2.14,18.1.5,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DebugIP *DebugIPAddr `json:"debug_ip,omitempty"`

	// Enables the use of kdump on SE. Requires SE Reboot. Applicable only in case of VM Based deployments. Field introduced in 18.2.10, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableKdump *bool `json:"enable_kdump,omitempty"`

	// Enable profiling time for certain RPC calls. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableRPCTimingProfiler *bool `json:"enable_rpc_timing_profiler,omitempty"`

	// Params for SE fault injection. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fault *DebugSeFault `json:"fault,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Flags []*DebugSeDataplane `json:"flags,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeagentDebug []*DebugSeAgent `json:"seagent_debug,omitempty"`

	// Debug knob for se_log_agent process. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SelogagentDebug *DebugSeAgent `json:"selogagent_debug,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Trace the functions calling memory allocation and free APIs. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TraceMemory *DebugTraceMemory `json:"trace_memory,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}

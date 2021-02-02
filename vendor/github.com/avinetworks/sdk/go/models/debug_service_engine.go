package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugServiceEngine debug service engine
// swagger:model DebugServiceEngine
type DebugServiceEngine struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Action to be invoked at configured layer. Enum options - SE_BENCHMARK_MODE_DROP, SE_BENCHMARK_MODE_REFLECT. Field introduced in 20.1.3.
	BenchmarkAction *string `json:"benchmark_action,omitempty"`

	// Toggle and configure the layer to benchmark performance. This can be done at a specific point in the SE packet processing pipeline. Enum options - SE_BENCHMARK_LAYER_NONE, SE_BENCHMARK_LAYER_POST_VNIC_RX, SE_BENCHMARK_LAYER_POST_FT_LOOKUP, SE_BENCHMARK_LAYER_NSP_LOOKUP, SE_BENCHMARK_LAYER_PRE_PROXY_PUNT, SE_BENCHMARK_LAYER_POST_PROXY_PUNT, SE_BENCHMARK_LAYER_ETHER_INPUT, SE_BENCHMARK_LAYER_IP_INPUT, SE_BENCHMARK_LAYER_UDP_INPUT. Field introduced in 20.1.3.
	BenchmarkLayer *string `json:"benchmark_layer,omitempty"`

	// Configure different reflect modes. Enum options - SE_BENCHMARK_REFLECT_SWAP_L4, SE_BENCHMARK_REFLECT_SWAP_L2, SE_BENCHMARK_REFLECT_SWAP_L3. Field introduced in 20.1.3.
	BenchmarkOption *string `json:"benchmark_option,omitempty"`

	// RSS Hash function to be used for packet reflect in TX path. Enum options - SE_BENCHMARK_DISABLE_HASH, SE_BENCHMARK_RTE_SOFT_HASH. Field introduced in 20.1.3.
	BenchmarkRssHash *string `json:"benchmark_rss_hash,omitempty"`

	// Enable/disable packet capture. Field introduced in 18.2.2.
	Capture *bool `json:"capture,omitempty"`

	// Per packet capture filters for Debug Service Engine. Not applicable for DOS pcap capture. . Field introduced in 18.2.5.
	CaptureFilters *CaptureFilters `json:"capture_filters,omitempty"`

	// Params for SE pcap. Field introduced in 17.2.14,18.1.5,18.2.1.
	CaptureParams *DebugVirtualServiceCapture `json:"capture_params,omitempty"`

	// Placeholder for description of property cpu_shares of obj type DebugServiceEngine field type str  type object
	CPUShares []*DebugSeCPUShares `json:"cpu_shares,omitempty"`

	// Per packet IP filter for Service Engine PCAP. Matches with source and destination address. Field introduced in 17.2.14,18.1.5,18.2.1.
	DebugIP *DebugIPAddr `json:"debug_ip,omitempty"`

	// Enables the use of kdump on SE. Requires SE Reboot. Applicable only in case of VM Based deployments. Field introduced in 18.2.10, 20.1.1.
	EnableKdump *bool `json:"enable_kdump,omitempty"`

	// Params for SE fault injection. Field introduced in 18.1.2.
	Fault *DebugSeFault `json:"fault,omitempty"`

	// Placeholder for description of property flags of obj type DebugServiceEngine field type str  type object
	Flags []*DebugSeDataplane `json:"flags,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Placeholder for description of property seagent_debug of obj type DebugServiceEngine field type str  type object
	SeagentDebug []*DebugSeAgent `json:"seagent_debug,omitempty"`

	// Debug knob for se_log_agent process. Field introduced in 20.1.1.
	SelogagentDebug *DebugSeAgent `json:"selogagent_debug,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}

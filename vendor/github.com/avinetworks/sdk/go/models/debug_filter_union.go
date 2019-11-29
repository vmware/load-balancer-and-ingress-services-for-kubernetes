package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugFilterUnion debug filter union
// swagger:model DebugFilterUnion
type DebugFilterUnion struct {

	// Placeholder for description of property alert_debug_filter of obj type DebugFilterUnion field type str  type object
	AlertDebugFilter *AlertMgrDebugFilter `json:"alert_debug_filter,omitempty"`

	// Placeholder for description of property autoscale_mgr_debug_filter of obj type DebugFilterUnion field type str  type object
	AutoscaleMgrDebugFilter *AutoScaleMgrDebugFilter `json:"autoscale_mgr_debug_filter,omitempty"`

	// Placeholder for description of property cloud_connector_debug_filter of obj type DebugFilterUnion field type str  type object
	CloudConnectorDebugFilter *CloudConnectorDebugFilter `json:"cloud_connector_debug_filter,omitempty"`

	// Placeholder for description of property hs_debug_filter of obj type DebugFilterUnion field type str  type object
	HsDebugFilter *HSMgrDebugFilter `json:"hs_debug_filter,omitempty"`

	// Placeholder for description of property mesos_metrics_debug_filter of obj type DebugFilterUnion field type str  type object
	MesosMetricsDebugFilter *MesosMetricsDebugFilter `json:"mesos_metrics_debug_filter,omitempty"`

	// Placeholder for description of property metrics_debug_filter of obj type DebugFilterUnion field type str  type object
	MetricsDebugFilter *MetricsMgrDebugFilter `json:"metrics_debug_filter,omitempty"`

	// Placeholder for description of property se_mgr_debug_filter of obj type DebugFilterUnion field type str  type object
	SeMgrDebugFilter *SeMgrDebugFilter `json:"se_mgr_debug_filter,omitempty"`

	// Add SE RPC Proxy Filter. Field introduced in 18.1.5, 18.2.1.
	SeRPCProxyFilter *SeRPCProxyDebugFilter `json:"se_rpc_proxy_filter,omitempty"`

	// Placeholder for description of property state_cache_mgr_debug_filter of obj type DebugFilterUnion field type str  type object
	StateCacheMgrDebugFilter *StateCacheMgrDebugFilter `json:"state_cache_mgr_debug_filter,omitempty"`

	//  Enum options - TASK_QUEUE_DEBUG, RPC_INFRA_DEBUG, JOB_MGR_DEBUG, TRANSACTION_DEBUG, SE_AGENT_DEBUG, SE_AGENT_METRICS_DEBUG, VIRTUALSERVICE_DEBUG, RES_MGR_DEBUG, SE_MGR_DEBUG, VI_MGR_DEBUG, METRICS_MANAGER_DEBUG, METRICS_MGR_DEBUG, EVENT_API_DEBUG, HS_MGR_DEBUG, ALERT_MGR_DEBUG, AUTOSCALE_MGR_DEBUG, APIC_AGENT_DEBUG, REDIS_INFRA_DEBUG, CLOUD_CONNECTOR_DEBUG, MESOS_METRICS_DEBUG, STATECACHE_MGR_DEBUG, NSX_AGENT_DEBUG, SE_AGENT_CPU_UTIL_DEBUG, SE_AGENT_MEM_UTIL_DEBUG, SE_RPC_PROXY_DEBUG.
	// Required: true
	Type *string `json:"type"`

	// Placeholder for description of property vs_debug_filter of obj type DebugFilterUnion field type str  type object
	VsDebugFilter *VsDebugFilter `json:"vs_debug_filter,omitempty"`
}

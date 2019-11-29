package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugSeAgent debug se agent
// swagger:model DebugSeAgent
type DebugSeAgent struct {

	// Log every nth message. Field introduced in 17.2.7.
	LogEveryN *int32 `json:"log_every_n,omitempty"`

	//  Enum options - LOG_LEVEL_DISABLED, LOG_LEVEL_INFO, LOG_LEVEL_WARNING, LOG_LEVEL_ERROR.
	// Required: true
	LogLevel *string `json:"log_level"`

	//  Enum options - TASK_QUEUE_DEBUG, RPC_INFRA_DEBUG, JOB_MGR_DEBUG, TRANSACTION_DEBUG, SE_AGENT_DEBUG, SE_AGENT_METRICS_DEBUG, VIRTUALSERVICE_DEBUG, RES_MGR_DEBUG, SE_MGR_DEBUG, VI_MGR_DEBUG, METRICS_MANAGER_DEBUG, METRICS_MGR_DEBUG, EVENT_API_DEBUG, HS_MGR_DEBUG, ALERT_MGR_DEBUG, AUTOSCALE_MGR_DEBUG, APIC_AGENT_DEBUG, REDIS_INFRA_DEBUG, CLOUD_CONNECTOR_DEBUG, MESOS_METRICS_DEBUG, STATECACHE_MGR_DEBUG, NSX_AGENT_DEBUG, SE_AGENT_CPU_UTIL_DEBUG, SE_AGENT_MEM_UTIL_DEBUG, SE_RPC_PROXY_DEBUG.
	// Required: true
	SubModule *string `json:"sub_module"`

	//  Enum options - TRACE_LEVEL_DISABLED, TRACE_LEVEL_ERROR, TRACE_LEVEL_DEBUG, TRACE_LEVEL_DEBUG_DETAIL.
	// Required: true
	TraceLevel *string `json:"trace_level"`
}

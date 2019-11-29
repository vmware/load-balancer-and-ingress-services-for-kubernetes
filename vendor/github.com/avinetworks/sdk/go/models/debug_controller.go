package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugController debug controller
// swagger:model DebugController
type DebugController struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property filters of obj type DebugController field type str  type object
	Filters *DebugFilterUnion `json:"filters,omitempty"`

	//  Enum options - LOG_LEVEL_DISABLED, LOG_LEVEL_INFO, LOG_LEVEL_WARNING, LOG_LEVEL_ERROR.
	// Required: true
	LogLevel *string `json:"log_level"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	//  Enum options - TASK_QUEUE_DEBUG, RPC_INFRA_DEBUG, JOB_MGR_DEBUG, TRANSACTION_DEBUG, SE_AGENT_DEBUG, SE_AGENT_METRICS_DEBUG, VIRTUALSERVICE_DEBUG, RES_MGR_DEBUG, SE_MGR_DEBUG, VI_MGR_DEBUG, METRICS_MANAGER_DEBUG, METRICS_MGR_DEBUG, EVENT_API_DEBUG, HS_MGR_DEBUG, ALERT_MGR_DEBUG, AUTOSCALE_MGR_DEBUG, APIC_AGENT_DEBUG, REDIS_INFRA_DEBUG, CLOUD_CONNECTOR_DEBUG, MESOS_METRICS_DEBUG, STATECACHE_MGR_DEBUG, NSX_AGENT_DEBUG, SE_AGENT_CPU_UTIL_DEBUG, SE_AGENT_MEM_UTIL_DEBUG, SE_RPC_PROXY_DEBUG.
	// Required: true
	SubModule *string `json:"sub_module"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Enum options - TRACE_LEVEL_DISABLED, TRACE_LEVEL_ERROR, TRACE_LEVEL_DEBUG, TRACE_LEVEL_DEBUG_DETAIL.
	// Required: true
	TraceLevel *string `json:"trace_level"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}

// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugFilterUnion debug filter union
// swagger:model DebugFilterUnion
type DebugFilterUnion struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AlertDebugFilter *AlertMgrDebugFilter `json:"alert_debug_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoscaleMgrDebugFilter *AutoScaleMgrDebugFilter `json:"autoscale_mgr_debug_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudConnectorDebugFilter *CloudConnectorDebugFilter `json:"cloud_connector_debug_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HsDebugFilter *HSMgrDebugFilter `json:"hs_debug_filter,omitempty"`

	// Add filter to Log Manager Debug. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogmanagerDebugFilter *LogManagerDebugFilter `json:"logmanager_debug_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MesosMetricsDebugFilter *MesosMetricsDebugFilter `json:"mesos_metrics_debug_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsDebugFilter *MetricsMgrDebugFilter `json:"metrics_debug_filter,omitempty"`

	// Add Metricsapi Server filter. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsapiSrvDebugFilter *MetricsAPISrvDebugFilter `json:"metricsapi_srv_debug_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMgrDebugFilter *SeMgrDebugFilter `json:"se_mgr_debug_filter,omitempty"`

	// Add SE RPC Proxy Filter. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRPCProxyFilter *SeRPCProxyDebugFilter `json:"se_rpc_proxy_filter,omitempty"`

	// Add Metricsapi Server filter. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecuritymgrDebugFilter *SecurityMgrDebugFilter `json:"securitymgr_debug_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StateCacheMgrDebugFilter *StateCacheMgrDebugFilter `json:"state_cache_mgr_debug_filter,omitempty"`

	//  Enum options - TASK_QUEUE_DEBUG, RPC_INFRA_DEBUG, JOB_MGR_DEBUG, TRANSACTION_DEBUG, SE_AGENT_DEBUG, SE_AGENT_METRICS_DEBUG, VIRTUALSERVICE_DEBUG, RES_MGR_DEBUG, SE_MGR_DEBUG, VI_MGR_DEBUG, METRICS_MANAGER_DEBUG, METRICS_MGR_DEBUG, EVENT_API_DEBUG, HS_MGR_DEBUG, ALERT_MGR_DEBUG, AUTOSCALE_MGR_DEBUG, APIC_AGENT_DEBUG, REDIS_INFRA_DEBUG, CLOUD_CONNECTOR_DEBUG, MESOS_METRICS_DEBUG, STATECACHE_MGR_DEBUG, NSX_AGENT_DEBUG, SE_AGENT_CPU_UTIL_DEBUG, SE_AGENT_MEM_UTIL_DEBUG, SE_RPC_PROXY_DEBUG, SE_AGENT_GSLB_DEBUG, METRICSAPI_SRV_DEBUG, SECURITYMGR_DEBUG, RES_MGR_READ_DEBUG, LICENSE_VMWSRVR_DEBUG, SE_AGENT_RESOLVERDB_DEBUG, LOGMANAGER_DEBUG, OSYNC_DEBUG. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsDebugFilter *VsDebugFilter `json:"vs_debug_filter,omitempty"`
}

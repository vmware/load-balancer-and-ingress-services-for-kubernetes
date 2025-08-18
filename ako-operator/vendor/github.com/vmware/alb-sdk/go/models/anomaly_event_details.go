// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AnomalyEventDetails anomaly event details
// swagger:model AnomalyEventDetails
type AnomalyEventDetails struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Deviation *float64 `json:"deviation,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MetricID *string `json:"metric_id"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MetricTimestamp *string `json:"metric_timestamp"`

	// Deprecated. Enum options - EXPONENTIAL_MOVING_AVG, EXPONENTIAL_WEIGHTED_MOVING_AVG, HOLTWINTERS_AT_AS, HOLTWINTERS_AT_MS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Model *string `json:"model,omitempty"`

	//  Enum options - EXPONENTIAL_MOVING_AVG, EXPONENTIAL_WEIGHTED_MOVING_AVG, HOLTWINTERS_AT_AS, HOLTWINTERS_AT_MS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Models []string `json:"models,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeID *string `json:"node_id,omitempty"`

	//  Enum options - METRICS_OBJ_TYPE_UNKNOWN, VSERVER_L4_SERVER, VSERVER_L4_CLIENT, VSERVER_L7_SERVER, VSERVER_L7_CLIENT, VM_METRICS_OBJ, SE_METRICS_OBJ, VSERVER_RUM, CONTROLLER_METRICS_OBJ, METRICS_COLLECTION, METRICS_RUM_PREAGG_BROWSER_OBJ, METRICS_RUM_PREAGG_COUNTRY_OBJ, METRICS_RUM_PREAGG_DEVTYPE_OBJ, METRICS_RUM_PREAGG_LANG_OBJ, METRICS_RUM_PREAGG_OS_OBJ, METRICS_RUM_PREAGG_URL_OBJ, METRICS_ANOMALY_OBJ, METRICS_HEALTHSCORE_OBJ, METRICS_RESOURCE_TIMING_BROWSER_OBJ, METRICS_RESOURCE_TIMING_OS_OBJ, METRICS_RESOURCE_TIMING_COUNTRY_OBJ, METRICS_RESOURCE_TIMING_LANG_OBJ, METRICS_RESOURCE_TIMING_DEVTYPE_OBJ, METRICS_RESOURCE_TIMING_URL_OBJ, METRICS_RESOURCE_TIMING_DIMENSION_OBJ, METRICS_RESOURCE_TIMING_BLOB_OBJ, METRICS_DOS_OBJ, METRICS_RUM_PREAGG_IPGROUP_OBJ, METRICS_APP_INSIGHTS_OBJ, METRICS_VSERVER_DNS_OBJ, METRICS_SERVER_DNS_OBJ, METRICS_SERVICE_INSIGHTS_OBJ, METRICS_SOURCE_INSIGHTS_OBJ, METRICS_TENANT_STATS_OBJ, METRICS_SE_IF_STATS_OBJ, METRICS_USER_METRICS_OBJ, METRICS_WAF_GROUP_OBJ, METRICS_WAF_RULE_OBJ, METRICS_WAF_TAG_OBJ, METRICS_PROCESS_STATS_OBJ, METRICS_VSERVER_HTTP2_CLIENT_OBJ, METRICS_WAF_WHITELIST_OBJ, METRICS_WAF_PSM_GROUP_OBJ, METRICS_WAF_PSMLOCATION_OBJ, METRICS_WAF_PSM_RULE_OBJ, METRICS_PG_STAT_DATABASE_OBJ, METRICS_PG_STAT_ALL_TABLES_OBJ, METRICS_PG_STAT_ALL_INDEXES_OBJ, METRICS_PG_STAT_IO_ALL_TABLES_OBJ, METRICS_PG_STAT_CLASS_OBJ, METRICS_PG_STAT_BG_WRITER_OBJ, METRICS_GSLB_STATS_OBJ, METRICS_VS_SCALEOUT_OBJ, METRICS_API_PERF_STATS_OBJ, METRICS_NSXT_STATS_OBJ, METRICS_ICAP_OBJ, METRICS_BOT_OBJ, METRICS_SEGROUP_OBJ, ENVOY_UPSTREAM_STATS_OBJ, ENVOY_DOWNSTREAM_STATS_OBJ, REDIS_QUEUE_STATS_OBJ. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjType *string `json:"obj_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolName *string `json:"pool_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	//  Enum options - ANZ_PRIORITY_HIGH, ANZ_PRIORITY_MEDIUM, ANZ_PRIORITY_LOW. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Priority *string `json:"priority"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Server *string `json:"server,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Value *float64 `json:"value"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsDerivationData metrics derivation data
// swagger:model MetricsDerivationData
type MetricsDerivationData struct {

	//  Enum options - METRICS_ALIAS. SUM_FIRST_N_DIVIDE_BY_LAST. SUM_BW_GAUGE. AVG_GET_POST_OTHER_LATENCY. APPDEX_ON_3_BUCKETS. APPDEX_ON_4_BUCKETS. SUM_GAUGE. SUM_N_METRICS. APPDEX_ON_5_BUCKETS. APPDEX_ON_6_BUCKETS. APPDEX_ON_CONNECTIONS. APPDEX_ON_2_BUCKETS. AVG_CLIENT_LATENCY. AVG_APPLICATION_LATENCY. MIN_N_METRICS. SUM_FIRST_N_DIVIDE_BY_LAST_PERCENTAGE. L4_CONNECTION_ERROR_PERCENTAGE. AVG_L4_CLIENT_LATENCY. CHECK_FOR_TRANSITIONS. SUBSTRACT_ALL_FROM_FIRST...
	// Required: true
	DerivationFn *string `json:"derivation_fn"`

	// Placeholder for description of property exclude_derived_metric of obj type MetricsDerivationData field type str  type boolean
	ExcludeDerivedMetric *bool `json:"exclude_derived_metric,omitempty"`

	// Placeholder for description of property include_derivation_metrics of obj type MetricsDerivationData field type str  type boolean
	IncludeDerivationMetrics *bool `json:"include_derivation_metrics,omitempty"`

	//  Enum options - METRICS_TABLE_NONE. METRICS_TABLE_ANOMALY. METRICS_TABLE_CONTROLLER_STATS. METRICS_TABLE_HEALTH_SCORE. METRICS_TABLE_SE_STATS. METRICS_TABLE_VSERVER_L4_SERVER. METRICS_TABLE_VSERVER_L4_CLIENT. METRICS_TABLE_VSERVER_L7_CLIENT. METRICS_TABLE_VSERVER_L7_SERVER. METRICS_TABLE_RUM_PREAGG_BROWSER. METRICS_TABLE_RUM_PREAGG_COUNTRY. METRICS_TABLE_RUM_PREAGG_DEVTYPE. METRICS_TABLE_RUM_PREAGG_LANG. METRICS_TABLE_RUM_PREAGG_OS. METRICS_TABLE_RUM_PREAGG_URL. METRICS_TABLE_RUM_ANALYTICS. METRICS_TABLE_VM_STATS. METRICS_TABLE_RESOURCE_TIMING_DIM. METRICS_TABLE_RESOURCE_TIMING_BLOB. METRICS_TABLE_RUM_PREAGG_IPGROUP...
	JoinTables *string `json:"join_tables,omitempty"`

	// metric_ids of MetricsDerivationData.
	// Required: true
	MetricIds *string `json:"metric_ids"`

	// Placeholder for description of property result_has_additional_fields of obj type MetricsDerivationData field type str  type boolean
	ResultHasAdditionalFields *bool `json:"result_has_additional_fields,omitempty"`

	//  Field introduced in 17.2.8.
	SecondOrderDerivation *bool `json:"second_order_derivation,omitempty"`

	// Placeholder for description of property skip_backend_derivation of obj type MetricsDerivationData field type str  type boolean
	SkipBackendDerivation *bool `json:"skip_backend_derivation,omitempty"`
}

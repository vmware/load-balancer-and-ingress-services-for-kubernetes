package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AnalyticsPolicy analytics policy
// swagger:model AnalyticsPolicy
type AnalyticsPolicy struct {

	// Log all headers. Field introduced in 18.1.4, 18.2.1.
	AllHeaders *bool `json:"all_headers,omitempty"`

	// Gain insights from sampled client to server HTTP requests and responses. Enum options - NO_INSIGHTS, PASSIVE, ACTIVE.
	ClientInsights *string `json:"client_insights,omitempty"`

	// Placeholder for description of property client_insights_sampling of obj type AnalyticsPolicy field type str  type object
	ClientInsightsSampling *ClientInsightsSampling `json:"client_insights_sampling,omitempty"`

	// Placeholder for description of property client_log_filters of obj type AnalyticsPolicy field type str  type object
	ClientLogFilters []*ClientLogFilter `json:"client_log_filters,omitempty"`

	// [DEPRECATED] Disable Analytics on this VirtualService. This will disable the collection of both metrics and logs. Please use following fields in AnalytcsProfile to control this behavior instead. disable_vs_analytics (for VirtualServices metrics), disable_server_analytics (for Pool metrics) and client_log_config (for logs). Field deprecated in 18.2.1. Field introduced in 17.2.4.
	Enabled *bool `json:"enabled,omitempty"`

	// Placeholder for description of property full_client_logs of obj type AnalyticsPolicy field type str  type object
	FullClientLogs *FullClientLogs `json:"full_client_logs,omitempty"`

	// Settings to turn on realtime metrics and set duration for realtime updates.
	MetricsRealtimeUpdate *MetricsRealTimeUpdate `json:"metrics_realtime_update,omitempty"`

	// This setting limits the number of significant logs generated per second for this VS on each SE. Default is 10 logs per second. Set it to zero (0) to disable throttling. Field introduced in 17.1.3.
	SignificantLogThrottle *int32 `json:"significant_log_throttle,omitempty"`

	// This setting limits the total number of UDF logs generated per second for this VS on each SE. UDF logs are generated due to the configured client log filters or the rules with logging enabled. Default is 10 logs per second. Set it to zero (0) to disable throttling. Field introduced in 17.1.3.
	UdfLogThrottle *int32 `json:"udf_log_throttle,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PoolAnalyticsPolicy pool analytics policy
// swagger:model PoolAnalyticsPolicy
type PoolAnalyticsPolicy struct {

	// Enable real time metrics for server and pool metrics eg. l4_server.xxx, l7_server.xxx. Field introduced in 18.1.5, 18.2.1.
	EnableRealtimeMetrics *bool `json:"enable_realtime_metrics,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertRuleMetric alert rule metric
// swagger:model AlertRuleMetric
type AlertRuleMetric struct {

	// Evaluation window for the Metrics.
	Duration *int32 `json:"duration,omitempty"`

	// Metric Id for the Alert. Eg. l4_client.avg_complete_conns.
	MetricID *string `json:"metric_id,omitempty"`

	// Placeholder for description of property metric_threshold of obj type AlertRuleMetric field type str  type object
	// Required: true
	MetricThreshold *AlertMetricThreshold `json:"metric_threshold"`
}

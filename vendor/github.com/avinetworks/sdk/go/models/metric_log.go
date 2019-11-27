package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricLog metric log
// swagger:model MetricLog
type MetricLog struct {

	// Placeholder for description of property end_timestamp of obj type MetricLog field type str  type number
	EndTimestamp *float64 `json:"end_timestamp,omitempty"`

	// metric_id of MetricLog.
	// Required: true
	MetricID *string `json:"metric_id"`

	// Placeholder for description of property report_timestamp of obj type MetricLog field type str  type number
	ReportTimestamp *float64 `json:"report_timestamp,omitempty"`

	// Number of step.
	Step *int32 `json:"step,omitempty"`

	// Placeholder for description of property time_series of obj type MetricLog field type str  type object
	TimeSeries *MetricsQueryResponse `json:"time_series,omitempty"`

	// Placeholder for description of property value of obj type MetricLog field type str  type number
	// Required: true
	Value *float64 `json:"value"`
}

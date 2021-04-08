package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AvgUptimeChangeDetails avg uptime change details
// swagger:model AvgUptimeChangeDetails
type AvgUptimeChangeDetails struct {

	// Placeholder for description of property current_value of obj type AvgUptimeChangeDetails field type str  type number
	CurrentValue *float64 `json:"current_value,omitempty"`

	// metric_id of AvgUptimeChangeDetails.
	MetricID *string `json:"metric_id,omitempty"`

	// metric_name of AvgUptimeChangeDetails.
	MetricName *string `json:"metric_name,omitempty"`

	// resource_str of AvgUptimeChangeDetails.
	ResourceStr *string `json:"resource_str,omitempty"`

	// Placeholder for description of property threshold of obj type AvgUptimeChangeDetails field type str  type number
	Threshold *float64 `json:"threshold,omitempty"`
}

package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricThresoldUpDetails metric thresold up details
// swagger:model MetricThresoldUpDetails
type MetricThresoldUpDetails struct {

	// Placeholder for description of property current_value of obj type MetricThresoldUpDetails field type str  type number
	CurrentValue *float64 `json:"current_value,omitempty"`

	// ID of the object whose metric has hit the threshold.
	EntityUUID *string `json:"entity_uuid,omitempty"`

	// metric_id of MetricThresoldUpDetails.
	MetricID *string `json:"metric_id,omitempty"`

	// metric_name of MetricThresoldUpDetails.
	// Required: true
	MetricName *string `json:"metric_name"`

	// Identity of the Pool.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Server IP Port on which event was generated.
	Server *string `json:"server,omitempty"`

	// Placeholder for description of property threshold of obj type MetricThresoldUpDetails field type str  type number
	Threshold *float64 `json:"threshold,omitempty"`
}

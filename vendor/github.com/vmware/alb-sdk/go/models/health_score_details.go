package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthScoreDetails health score details
// swagger:model HealthScoreDetails
type HealthScoreDetails struct {

	// Number of anomaly_penalty.
	AnomalyPenalty *int32 `json:"anomaly_penalty,omitempty"`

	// Reason for Anomaly Penalty.
	AnomalyReason *string `json:"anomaly_reason,omitempty"`

	// Reason for Performance Score.
	PerformanceReason *string `json:"performance_reason,omitempty"`

	// Number of performance_score.
	PerformanceScore *int32 `json:"performance_score,omitempty"`

	// Placeholder for description of property previous_value of obj type HealthScoreDetails field type str  type number
	// Required: true
	PreviousValue *float64 `json:"previous_value"`

	// Reason for the Health Score Change.
	Reason *string `json:"reason,omitempty"`

	// Number of resources_penalty.
	ResourcesPenalty *int32 `json:"resources_penalty,omitempty"`

	// Reason for Resources Penalty.
	ResourcesReason *string `json:"resources_reason,omitempty"`

	// Number of security_penalty.
	SecurityPenalty *int32 `json:"security_penalty,omitempty"`

	// Reason for Security Threat Level.
	SecurityReason *string `json:"security_reason,omitempty"`

	// The step interval in seconds.
	Step *int32 `json:"step,omitempty"`

	// Resource prefix containing entity information.
	SubResourcePrefix *string `json:"sub_resource_prefix,omitempty"`

	// timestamp of HealthScoreDetails.
	// Required: true
	Timestamp *string `json:"timestamp"`

	// Placeholder for description of property value of obj type HealthScoreDetails field type str  type number
	// Required: true
	Value *float64 `json:"value"`
}

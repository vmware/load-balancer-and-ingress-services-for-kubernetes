package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertMetricThreshold alert metric threshold
// swagger:model AlertMetricThreshold
type AlertMetricThreshold struct {

	//  Enum options - ALERT_OP_LT, ALERT_OP_LE, ALERT_OP_EQ, ALERT_OP_NE, ALERT_OP_GE, ALERT_OP_GT.
	// Required: true
	Comparator *string `json:"comparator"`

	// Metric threshold for comparison.
	Threshold *int32 `json:"threshold,omitempty"`
}

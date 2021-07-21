// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertRuleMetric alert rule metric
// swagger:model AlertRuleMetric
type AlertRuleMetric struct {

	// Evaluation window for the Metrics. Unit is SEC.
	Duration *int32 `json:"duration,omitempty"`

	// Metric Id for the Alert. Eg. l4_client.avg_complete_conns.
	MetricID *string `json:"metric_id,omitempty"`

	// Placeholder for description of property metric_threshold of obj type AlertRuleMetric field type str  type object
	// Required: true
	MetricThreshold *AlertMetricThreshold `json:"metric_threshold"`
}

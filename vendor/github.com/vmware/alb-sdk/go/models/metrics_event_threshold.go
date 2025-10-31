// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsEventThreshold metrics event threshold
// swagger:model MetricsEventThreshold
type MetricsEventThreshold struct {

	// Type of the metrics event threshold. This value will decide which metric rule (or rules) use configured thresholds. Enum options - THRESHOLD_TYPE_STATIC, SE_CPU_THRESHOLD, SE_MEM_THRESHOLD, SE_DISK_THRESHOLD, CONTROLLER_CPU_THRESHOLD, CONTROLLER_MEM_THRESHOLD, CONTROLLER_DISK_THRESHOLD. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	MetricsEventThresholdType *string `json:"metrics_event_threshold_type"`

	// This value is used to reset the event state machine. Allowed values are 1-100. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResetThreshold *float64 `json:"reset_threshold,omitempty"`

	// Threshold value for which event in raised. There can be multiple thresholds defined.Health score degrades when the the target is higher than this threshold. Allowed values are 1-100. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	WatermarkThresholds []int64 `json:"watermark_thresholds,omitempty,omitempty"`
}

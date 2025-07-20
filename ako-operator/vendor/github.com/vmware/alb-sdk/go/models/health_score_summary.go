// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthScoreSummary health score summary
// swagger:model HealthScoreSummary
type HealthScoreSummary struct {

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AnomalyPenalty *uint32 `json:"anomaly_penalty,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HealthScore *float64 `json:"health_score,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PerformanceScore *HealthScorePerformanceData `json:"performance_score,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResourcesPenalty *uint32 `json:"resources_penalty,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SecurityPenalty *uint32 `json:"security_penalty,omitempty"`
}

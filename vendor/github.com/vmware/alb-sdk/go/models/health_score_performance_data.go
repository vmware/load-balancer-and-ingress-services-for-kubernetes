// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthScorePerformanceData health score performance data
// swagger:model HealthScorePerformanceData
type HealthScorePerformanceData struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ApplicationPerformance *ApplicationPerformanceScoreData `json:"application_performance,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolPerformance *PoolPerformanceScoreData `json:"pool_performance,omitempty"`

	// Reason for Performance Score. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Reason *string `json:"reason"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReasonAttr *string `json:"reason_attr,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerPerformance *ServerPerformanceScoreData `json:"server_performance,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceenginePerformance *ServiceEnginePerformanceScoreData `json:"serviceengine_performance,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Value *float64 `json:"value"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualservicePerformance *VirtualServicePerformanceScoreData `json:"virtualservice_performance,omitempty"`
}

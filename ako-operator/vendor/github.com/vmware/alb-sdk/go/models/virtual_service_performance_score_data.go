// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VirtualServicePerformanceScoreData virtual service performance score data
// swagger:model VirtualServicePerformanceScoreData
type VirtualServicePerformanceScoreData struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Apdexc *float64 `json:"apdexc,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Apdexr *float64 `json:"apdexr,omitempty"`

	// Average of all pool performance scores. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgPoolPerformanceScore *float64 `json:"avg_pool_performance_score,omitempty"`

	//  Enum options - OPER_UP, OPER_DOWN, OPER_CREATING, OPER_RESOURCES, OPER_INACTIVE, OPER_DISABLED, OPER_UNUSED, OPER_UNKNOWN, OPER_PROCESSING, OPER_INITIALIZING, OPER_ERROR_DISABLED, OPER_AWAIT_MANUAL_PLACEMENT, OPER_UPGRADING, OPER_SE_PROCESSING, OPER_PARTITIONED, OPER_DISABLING, OPER_FAILED, OPER_UNAVAIL, OPER_AGGREGATE_DOWN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	OperState *string `json:"oper_state"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolPerformanceScores []*PoolPerformanceScore `json:"pool_performance_scores,omitempty"`

	// Reason for the Health Score. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Reason *string `json:"reason"`

	// Attribute that is dominating the health score. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReasonAttr *string `json:"reason_attr,omitempty"`

	//  It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ref *string `json:"ref,omitempty"`

	// Rum Apdexr when client insights is active. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RumApdexr *float64 `json:"rum_apdexr,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SumFinishedConns *float64 `json:"sum_finished_conns,omitempty"`

	// Percentage time of last 5mins that the VirtualService has been up. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUptime *float64 `json:"vs_uptime,omitempty"`
}

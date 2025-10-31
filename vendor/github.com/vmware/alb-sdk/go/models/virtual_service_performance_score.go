// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VirtualServicePerformanceScore virtual service performance score
// swagger:model VirtualServicePerformanceScore
type VirtualServicePerformanceScore struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	HsEntity *HealthScoreEntity `json:"hs_entity"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Reason *string `json:"reason"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScoreData *VirtualServicePerformanceScoreData `json:"score_data,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Value *float64 `json:"value"`
}

// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ApplicationPerformanceScoreData application performance score data
// swagger:model ApplicationPerformanceScoreData
type ApplicationPerformanceScoreData struct {

	// Reason for the Health Score. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Reason *string `json:"reason"`

	// Attribute that is dominating the health score. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReasonAttr *string `json:"reason_attr,omitempty"`

	//  It is a reference to an object of type Application. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ref *string `json:"ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualservicePerformanceScores []*VirtualServicePerformanceScore `json:"virtualservice_performance_scores,omitempty"`
}

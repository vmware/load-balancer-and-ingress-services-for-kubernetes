// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PGDeploymentRuleResult p g deployment rule result
// swagger:model PGDeploymentRuleResult
type PGDeploymentRuleResult struct {

	// Metric value that is used as the pass fail. If it is not provided then it will simply compare it with current pool vs new pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricValue *float64 `json:"metric_value,omitempty"`

	// Whether rule passed or failed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Result *bool `json:"result,omitempty"`

	// Rule used for evaluation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Rule *PGDeploymentRule `json:"rule"`
}

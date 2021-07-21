// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PGDeploymentRule p g deployment rule
// swagger:model PGDeploymentRule
type PGDeploymentRule struct {

	// metric_id of PGDeploymentRule.
	MetricID *string `json:"metric_id,omitempty"`

	//  Enum options - CO_EQ, CO_GT, CO_GE, CO_LT, CO_LE, CO_NE.
	Operator *string `json:"operator,omitempty"`

	// metric threshold that is used as the pass fail. If it is not provided then it will simply compare it with current pool vs new pool.
	Threshold *float64 `json:"threshold,omitempty"`
}

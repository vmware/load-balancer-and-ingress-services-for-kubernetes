// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsMissingDataInterval metrics missing data interval
// swagger:model MetricsMissingDataInterval
type MetricsMissingDataInterval struct {

	// end of MetricsMissingDataInterval.
	// Required: true
	End *string `json:"end"`

	// start of MetricsMissingDataInterval.
	// Required: true
	Start *string `json:"start"`
}
